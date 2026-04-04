package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	sharedtypes "crona/shared/types"
)

func PrintJSON(w io.Writer, value any) error {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(body))
	return err
}

func OptionalFlag(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func OptionalText(value *string, fallback string) string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return fallback
	}
	return strings.TrimSpace(*value)
}

func PrintOK(w io.Writer) error {
	_, err := fmt.Fprintln(w, "ok")
	return err
}

func PrintContext(w io.Writer, out sharedtypes.ActiveContext) error {
	_, err := fmt.Fprintf(w, "repo: %s\nstream: %s\nissue: %s\n", OptionalText(out.RepoName, "-"), OptionalText(out.StreamName, "-"), OptionalText(out.IssueTitle, "-"))
	return err
}

func PrintContextInline(w io.Writer, prefix string, out sharedtypes.ActiveContext) error {
	_, err := fmt.Fprintf(w, "%s: repo=%s stream=%s issue=%s\n", prefix, OptionalText(out.RepoName, "-"), OptionalText(out.StreamName, "-"), OptionalText(out.IssueTitle, "-"))
	return err
}

func PrintTimerStatus(w io.Writer, out sharedtypes.TimerState) error {
	segment := "-"
	if out.SegmentType != nil {
		segment = string(*out.SegmentType)
	}
	issue := "-"
	if out.IssueID != nil {
		issue = fmt.Sprintf("%d", *out.IssueID)
	}
	session := "-"
	if out.SessionID != nil {
		session = *out.SessionID
	}
	_, err := fmt.Fprintf(w, "state: %s\nsegment: %s\nissue: %s\nsession: %s\nelapsed: %ds\n", out.State, segment, issue, session, out.ElapsedSeconds)
	return err
}

func PrintTimerResult(w io.Writer, out sharedtypes.TimerState, message string) error {
	_, err := fmt.Fprintf(w, "%s\nstate: %s\nelapsed: %ds\n", message, out.State, out.ElapsedSeconds)
	return err
}

func PrintKernelAttach(w io.Writer, info *sharedtypes.KernelInfo, endpoint string) error {
	_, err := fmt.Fprintf(w, "kernel attached\npid: %d\nendpoint: %s\n", info.PID, endpoint)
	return err
}

func PrintKernelInfo(w io.Writer, out sharedtypes.KernelInfo, transport, endpoint string) error {
	_, err := fmt.Fprintf(w, "pid: %d\ntransport: %s\nendpoint: %s\nenv: %s\nstarted: %s\nscratch: %s\n", out.PID, transport, endpoint, out.Env, out.StartedAt, out.ScratchDir)
	return err
}

func PrintExportResult(w io.Writer, out sharedtypes.ExportReportResult) error {
	if _, err := fmt.Fprintf(w, "kind: %s\nformat: %s\noutput: %s\n", out.Kind, out.Format, out.OutputMode); err != nil {
		return err
	}
	if out.FilePath != nil && strings.TrimSpace(*out.FilePath) != "" {
		_, err := fmt.Fprintf(w, "file: %s\n", strings.TrimSpace(*out.FilePath))
		return err
	}
	content := strings.TrimSpace(out.Content)
	if content == "" {
		_, err := fmt.Fprintln(w, "content: (empty)")
		return err
	}
	if _, err := fmt.Fprintln(w, "content:"); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, content)
	return err
}

func ExportReportLabel(report sharedtypes.ExportReportFile) string {
	if label := strings.TrimSpace(report.DateLabel); label != "" {
		return label
	}
	if label := strings.TrimSpace(report.Date); label != "" {
		return label
	}
	if label := strings.TrimSpace(report.EndDate); label != "" {
		return label
	}
	return report.Name
}

func PrintUpdateStatus(w io.Writer, status sharedtypes.UpdateStatus) error {
	if _, err := fmt.Fprintf(w, "current: %s\nenabled: %t\nprompt: %t\n", status.CurrentVersion, status.Enabled, status.PromptEnabled); err != nil {
		return err
	}
	if strings.TrimSpace(status.CheckedAt) != "" {
		if _, err := fmt.Fprintf(w, "checked: %s\n", status.CheckedAt); err != nil {
			return err
		}
	}
	if strings.TrimSpace(status.LatestVersion) != "" {
		if _, err := fmt.Fprintf(w, "latest: %s\n", status.LatestVersion); err != nil {
			return err
		}
	}
	if strings.TrimSpace(status.ReleaseName) != "" {
		if _, err := fmt.Fprintf(w, "title: %s\n", status.ReleaseName); err != nil {
			return err
		}
	} else if summary := FirstReleaseSummary(status.ReleaseNotes); summary != "" {
		if _, err := fmt.Fprintf(w, "summary: %s\n", summary); err != nil {
			return err
		}
	}
	if strings.TrimSpace(status.ReleaseURL) != "" {
		if _, err := fmt.Fprintf(w, "release: %s\n", status.ReleaseURL); err != nil {
			return err
		}
	}
	if status.UpdateAvailable {
		if _, err := fmt.Fprintln(w, "update: available"); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintln(w, "update: none"); err != nil {
			return err
		}
	}
	if strings.TrimSpace(status.DismissedVersion) != "" {
		if _, err := fmt.Fprintf(w, "dismissed: %s\n", status.DismissedVersion); err != nil {
			return err
		}
	}
	if strings.TrimSpace(status.Error) != "" {
		if _, err := fmt.Fprintf(w, "error: %s\n", status.Error); err != nil {
			return err
		}
	}
	return nil
}

func PrintUpdateNotes(w io.Writer, status sharedtypes.UpdateStatus) error {
	if err := PrintUpdateStatus(w, status); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "release notes:"); err != nil {
		return err
	}
	notes := strings.TrimSpace(status.ReleaseNotes)
	if notes == "" {
		_, err := fmt.Fprintln(w, "  (no release notes published)")
		return err
	}
	_, err := fmt.Fprintln(w, notes)
	return err
}

func FirstReleaseSummary(body string) string {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
		if line != "" {
			return line
		}
	}
	return ""
}
