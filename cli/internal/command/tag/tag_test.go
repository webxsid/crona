package tag

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

type call struct {
	method string
	params any
}

func fakeDeps(out *bytes.Buffer, calls *[]call, reply any) Deps {
	return Deps{
		Stdout: out,
		CallKernel: func(method string, params, dst any) error {
			*calls = append(*calls, call{method, params})
			if dst != nil && reply != nil {
				raw, _ := json.Marshal(reply)
				return json.Unmarshal(raw, dst)
			}
			return nil
		},
	}
}

func TestListPrintsCountsAndKind(t *testing.T) {
	var out bytes.Buffer
	var calls []call
	red := "red"
	tags := []sharedtypes.Tag{
		{Name: "urgent", Color: &red, System: true, IssueCount: 3},
		{Name: "client-acme", HabitCount: 1},
	}
	if err := Run([]string{"ls"}, fakeDeps(&out, &calls, tags)); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{"+urgent", "red", "system", "+client-acme", "user"} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}
}

func TestRenameAcceptsMergeAfterNames(t *testing.T) {
	var out bytes.Buffer
	var calls []call
	if err := Run([]string{"rename", "old", "new", "--merge"}, fakeDeps(&out, &calls, nil)); err != nil {
		t.Fatal(err)
	}
	req := calls[0].params.(shareddto.TagRenameRequest)
	if calls[0].method != protocol.MethodTagRename || req.From != "old" || req.To != "new" || !req.Merge {
		t.Fatalf("unexpected call %+v", calls[0])
	}
}

func TestColorNoneClears(t *testing.T) {
	var out bytes.Buffer
	var calls []call
	if err := Run([]string{"color", "docs", "none"}, fakeDeps(&out, &calls, nil)); err != nil {
		t.Fatal(err)
	}
	if req := calls[0].params.(shareddto.TagSetColorRequest); req.Color != nil {
		t.Fatalf("expected nil color, got %v", *req.Color)
	}
}

func TestRunItemSplitsAddAndRemove(t *testing.T) {
	var out bytes.Buffer
	var calls []call
	if err := RunItem("habit", []string{"7", "+health", "-someday"}, fakeDeps(&out, &calls, map[string]any{})); err != nil {
		t.Fatal(err)
	}
	req := calls[0].params.(shareddto.UpdateTagsRequest)
	if calls[0].method != protocol.MethodHabitTagsUpdate || req.IDs[0] != 7 ||
		req.Add[0] != "health" || req.Remove[0] != "someday" {
		t.Fatalf("unexpected call %+v", calls[0])
	}
	if err := RunItem("issue", []string{"7", "urgent"}, fakeDeps(&out, &calls, nil)); err == nil {
		t.Fatal("expected an error for a tag without + or -")
	}
}
