package capture

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

type fakeKernel struct {
	parsed    []shareddto.IssueCaptureProposal
	parseReq  shareddto.IssueCaptureParseRequest
	imported  []shareddto.IssueCaptureProposal
	importErr error
}

func (k *fakeKernel) call(method string, params, dst any) error {
	switch method {
	case protocol.MethodIssueCaptureParse:
		k.parseReq = params.(shareddto.IssueCaptureParseRequest)
		raw, _ := json.Marshal(k.parsed)
		return json.Unmarshal(raw, dst)
	case protocol.MethodIssueCaptureImport:
		k.imported = params.(shareddto.IssueCaptureImportRequest).Proposals
		if k.importErr != nil {
			return k.importErr
		}
		out := dst.(*shareddto.IssueCaptureImportResult)
		out.Issues = make([]sharedtypes.Issue, len(k.imported))
		return nil
	}
	return errors.New("unexpected method " + method)
}

func proposals() []shareddto.IssueCaptureProposal {
	minutes := 45
	return []shareddto.IssueCaptureProposal{
		{Line: 1, Title: "fix login", RepoName: "crona", StreamName: "api", StreamID: 10, EstimateMinutes: &minutes, Tags: []string{"urgent"}},
		{Line: 2, Title: "broken", Flags: []string{"unknown #nope"}},
	}
}

func run(t *testing.T, k *fakeKernel, stdin string, confirm bool, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	err := Run(args, Deps{
		Stdout:     &out,
		Stdin:      strings.NewReader(stdin),
		CallKernel: k.call,
		Confirm:    func(string) (bool, error) { return confirm, nil },
	})
	return out.String(), err
}

func TestAddImportsUnflaggedRowsAfterConfirm(t *testing.T) {
	k := &fakeKernel{parsed: proposals()}
	out, err := run(t, k, "", true, "fix", "login", "#api", "--stream", "api", "--repo", "crona")
	if err != nil {
		t.Fatal(err)
	}
	if k.parseReq.Text != "- fix login #api" || k.parseReq.Target != "crona/api" {
		t.Fatalf("unexpected parse request %+v", k.parseReq)
	}
	if len(k.imported) != 1 || k.imported[0].Title != "fix login" {
		t.Fatalf("expected only the unflagged row imported, got %+v", k.imported)
	}
	for _, want := range []string{"crona/api", "45m", "+urgent", "?", "unknown #nope", "Imported 1 issue · 1 flagged row left out"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestAddDeclinedOrDryRunImportsNothing(t *testing.T) {
	k := &fakeKernel{parsed: proposals()}
	if out, err := run(t, k, "", false, "x"); err != nil || !strings.Contains(out, "Nothing imported") || k.imported != nil {
		t.Fatalf("declined run: %v %q %+v", err, out, k.imported)
	}
	out, err := run(t, k, "- a\n- b\n", true, "-", "--dry-run")
	if err != nil || k.imported != nil || k.parseReq.Text != "- a\n- b\n" || !strings.Contains(out, `"title": "fix login"`) {
		t.Fatalf("dry run: %v %q %+v", err, out, k.parseReq)
	}
}
