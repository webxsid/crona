package helpers

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	sharedtypes "crona/shared/types"
	"crona/tui/internal/api"
)

func SupportDiagnosticsSummary(info *api.KernelInfo, assets *api.ExportAssetStatus, update *api.UpdateStatus, health *api.Health, tuiPath, kernelPath string) string {
	lines := []string{}
	if update != nil {
		lines = append(lines, fmt.Sprintf("Version: v%s", fallbackSupport(strings.TrimSpace(update.CurrentVersion), "unknown")))
		lines = append(lines, fmt.Sprintf("Running channel: %s", fallbackSupport(strings.TrimSpace(string(update.RunningChannel)), "stable")))
		lines = append(lines, fmt.Sprintf("Update channel: %s", fallbackSupport(strings.TrimSpace(string(update.Channel)), "stable")))
	}
	if info != nil {
		lines = append(lines, fmt.Sprintf("Environment: %s", fallbackSupport(strings.TrimSpace(info.Env), "unknown")))
		lines = append(lines, fmt.Sprintf("Kernel transport: %s", fallbackSupport(strings.TrimSpace(info.Transport), "-")))
		lines = append(lines, fmt.Sprintf("Kernel endpoint: %s", fallbackSupport(strings.TrimSpace(info.Endpoint), "-")))
		lines = append(lines, fmt.Sprintf("Scratch dir: %s", fallbackSupport(strings.TrimSpace(info.ScratchDir), "-")))
	}
	if health != nil {
		lines = append(lines, fmt.Sprintf("Health: %s (db=%t)", fallbackSupport(strings.TrimSpace(health.Status), "unknown"), health.DB))
	}
	if assets != nil {
		lines = append(lines, fmt.Sprintf("Reports dir: %s", fallbackSupport(strings.TrimSpace(assets.ReportsDir), "-")))
		lines = append(lines, fmt.Sprintf("ICS dir: %s", fallbackSupport(strings.TrimSpace(assets.ICSDir), "-")))
	}
	lines = append(lines, fmt.Sprintf("TUI path: %s", fallbackSupport(strings.TrimSpace(tuiPath), "-")))
	lines = append(lines, fmt.Sprintf("Engine path: %s", fallbackSupport(strings.TrimSpace(kernelPath), "-")))
	return strings.Join(lines, "\n")
}

type SupportDiagnosticsInput struct {
	View                string
	Pane                string
	Width               int
	Height              int
	DashboardDate       string
	RollupStartDate     string
	RollupEndDate       string
	WellbeingDate       string
	ReposCount          int
	StreamsCount        int
	IssuesCount         int
	AllIssuesCount      int
	HabitsCount         int
	DueHabitsCount      int
	ReportsCount        int
	SessionHistoryCount int
	ScratchpadsCount    int
	OpsCount            int
	Context             *api.ActiveContext
	Timer               *api.TimerState
	Settings            *api.CoreSettings
	KernelInfo          *api.KernelInfo
	ExportAssets        *api.ExportAssetStatus
	UpdateStatus        *api.UpdateStatus
	Health              *api.Health
	TUIPath             string
	KernelPath          string
}

func SupportDiagnosticsReport(input SupportDiagnosticsInput) string {
	lines := []string{
		"Crona Diagnostics Report",
		fmt.Sprintf("Generated: %s", time.Now().UTC().Format(time.RFC3339)),
		fmt.Sprintf("Platform: %s/%s", runtime.GOOS, runtime.GOARCH),
		fmt.Sprintf("Go version: %s", runtime.Version()),
		"",
		"[app]",
		fmt.Sprintf("view=%s", fallbackSupport(input.View, "-")),
		fmt.Sprintf("pane=%s", fallbackSupport(input.Pane, "-")),
		fmt.Sprintf("size=%dx%d", input.Width, input.Height),
		fmt.Sprintf("dashboard_date=%s", fallbackSupport(input.DashboardDate, "-")),
		fmt.Sprintf("rollup_start=%s", fallbackSupport(input.RollupStartDate, "-")),
		fmt.Sprintf("rollup_end=%s", fallbackSupport(input.RollupEndDate, "-")),
		fmt.Sprintf("wellbeing_date=%s", fallbackSupport(input.WellbeingDate, "-")),
		"",
		"[runtime]",
		fmt.Sprintf("tui_path=%s", fallbackSupport(strings.TrimSpace(input.TUIPath), "-")),
		fmt.Sprintf("kernel_path=%s", fallbackSupport(strings.TrimSpace(input.KernelPath), "-")),
	}

	if input.UpdateStatus != nil {
		lines = append(lines,
			fmt.Sprintf("version=v%s", fallbackSupport(strings.TrimSpace(input.UpdateStatus.CurrentVersion), "unknown")),
			fmt.Sprintf("running_channel=%s", fallbackSupport(strings.TrimSpace(string(input.UpdateStatus.RunningChannel)), "stable")),
			fmt.Sprintf("running_is_beta=%t", input.UpdateStatus.RunningIsBeta),
			fmt.Sprintf("update_channel=%s", fallbackSupport(strings.TrimSpace(string(input.UpdateStatus.Channel)), "stable")),
			fmt.Sprintf("latest_version=%s", fallbackSupport(strings.TrimSpace(input.UpdateStatus.LatestVersion), "-")),
			fmt.Sprintf("release_tag=%s", fallbackSupport(strings.TrimSpace(input.UpdateStatus.ReleaseTag), "-")),
			fmt.Sprintf("release_prerelease=%t", input.UpdateStatus.ReleaseIsPrerelease),
			fmt.Sprintf("latest_is_beta=%t", input.UpdateStatus.LatestIsBeta),
			fmt.Sprintf("checked_at=%s", fallbackSupport(strings.TrimSpace(input.UpdateStatus.CheckedAt), "-")),
			fmt.Sprintf("update_available=%t", input.UpdateStatus.UpdateAvailable),
			fmt.Sprintf("install_available=%t", input.UpdateStatus.InstallAvailable),
			fmt.Sprintf("update_error=%s", fallbackSupport(strings.TrimSpace(input.UpdateStatus.Error), "-")),
		)
	}
	if input.KernelInfo != nil {
		lines = append(lines,
			fmt.Sprintf("env=%s", fallbackSupport(strings.TrimSpace(input.KernelInfo.Env), "unknown")),
			fmt.Sprintf("kernel_transport=%s", fallbackSupport(strings.TrimSpace(input.KernelInfo.Transport), "-")),
			fmt.Sprintf("kernel_endpoint=%s", fallbackSupport(strings.TrimSpace(input.KernelInfo.Endpoint), "-")),
			fmt.Sprintf("kernel_pid=%d", input.KernelInfo.PID),
			fmt.Sprintf("kernel_started=%s", fallbackSupport(strings.TrimSpace(input.KernelInfo.StartedAt), "-")),
			fmt.Sprintf("kernel_running_channel=%s", fallbackSupport(strings.TrimSpace(string(input.KernelInfo.RunningChannel)), "stable")),
			fmt.Sprintf("kernel_running_is_beta=%t", input.KernelInfo.RunningIsBeta),
			fmt.Sprintf("scratch_dir=%s", fallbackSupport(strings.TrimSpace(input.KernelInfo.ScratchDir), "-")),
		)
	}
	if input.Health != nil {
		lines = append(lines,
			fmt.Sprintf("health_status=%s", fallbackSupport(strings.TrimSpace(input.Health.Status), "unknown")),
			fmt.Sprintf("health_db=%t", input.Health.DB),
			fmt.Sprintf("health_uptime_seconds=%.0f", input.Health.Uptime),
		)
	}

	lines = append(lines, "", "[context]")
	if input.Context != nil {
		lines = append(lines,
			fmt.Sprintf("repo_id=%s", optionalInt64String(input.Context.RepoID)),
			fmt.Sprintf("repo_name=%s", fallbackSupport(pointerString(input.Context.RepoName), "-")),
			fmt.Sprintf("stream_id=%s", optionalInt64String(input.Context.StreamID)),
			fmt.Sprintf("stream_name=%s", fallbackSupport(pointerString(input.Context.StreamName), "-")),
			fmt.Sprintf("issue_id=%s", optionalInt64String(input.Context.IssueID)),
			fmt.Sprintf("issue_title=%s", fallbackSupport(pointerString(input.Context.IssueTitle), "-")),
			fmt.Sprintf("context_updated=%s", fallbackSupport(pointerString(input.Context.UpdatedAt), "-")),
		)
	} else {
		lines = append(lines, "context=unavailable")
	}

	lines = append(lines, "", "[timer]")
	if input.Timer != nil {
		lines = append(lines,
			fmt.Sprintf("state=%s", fallbackSupport(strings.TrimSpace(input.Timer.State), "unknown")),
			fmt.Sprintf("session_id=%s", fallbackSupport(pointerString(input.Timer.SessionID), "-")),
			fmt.Sprintf("issue_id=%s", optionalInt64String(input.Timer.IssueID)),
			fmt.Sprintf("segment_type=%s", fallbackSupport(optionalSegmentString(input.Timer.SegmentType), "-")),
			fmt.Sprintf("elapsed_seconds=%d", input.Timer.ElapsedSeconds),
		)
	} else {
		lines = append(lines, "timer=unavailable")
	}

	lines = append(lines, "", "[settings]")
	if input.Settings != nil {
		lines = append(lines,
			fmt.Sprintf("timer_mode=%s", input.Settings.TimerMode),
			fmt.Sprintf("breaks_enabled=%t", input.Settings.BreaksEnabled),
			fmt.Sprintf("work_minutes=%d", input.Settings.WorkDurationMinutes),
			fmt.Sprintf("short_break_minutes=%d", input.Settings.ShortBreakMinutes),
			fmt.Sprintf("long_break_minutes=%d", input.Settings.LongBreakMinutes),
			fmt.Sprintf("cycles_before_long_break=%d", input.Settings.CyclesBeforeLongBreak),
			fmt.Sprintf("update_checks_enabled=%t", input.Settings.UpdateChecksEnabled),
			fmt.Sprintf("update_prompt_enabled=%t", input.Settings.UpdatePromptEnabled),
			fmt.Sprintf("away_mode_enabled=%t", input.Settings.AwayModeEnabled),
			fmt.Sprintf("rollback_minutes=%d", input.Settings.DailyPlanRollbackMins),
		)
	} else {
		lines = append(lines, "settings=unavailable")
	}

	lines = append(lines, "", "[data_counts]",
		fmt.Sprintf("repos=%d", input.ReposCount),
		fmt.Sprintf("streams=%d", input.StreamsCount),
		fmt.Sprintf("issues=%d", input.IssuesCount),
		fmt.Sprintf("all_issues=%d", input.AllIssuesCount),
		fmt.Sprintf("habits=%d", input.HabitsCount),
		fmt.Sprintf("due_habits=%d", input.DueHabitsCount),
		fmt.Sprintf("session_history=%d", input.SessionHistoryCount),
		fmt.Sprintf("scratchpads=%d", input.ScratchpadsCount),
		fmt.Sprintf("ops=%d", input.OpsCount),
		fmt.Sprintf("export_reports=%d", input.ReportsCount),
	)

	lines = append(lines, "", "[export]")
	if input.ExportAssets != nil {
		lines = append(lines,
			fmt.Sprintf("reports_dir=%s", fallbackSupport(strings.TrimSpace(input.ExportAssets.ReportsDir), "-")),
			fmt.Sprintf("ics_dir=%s", fallbackSupport(strings.TrimSpace(input.ExportAssets.ICSDir), "-")),
			fmt.Sprintf("reports_dir_customized=%t", input.ExportAssets.ReportsDirCustomized),
			fmt.Sprintf("ics_dir_customized=%t", input.ExportAssets.ICSDirCustomized),
			fmt.Sprintf("pdf_renderer_available=%t", input.ExportAssets.PDFRendererAvailable),
			fmt.Sprintf("pdf_renderer_name=%s", fallbackSupport(strings.TrimSpace(input.ExportAssets.PDFRendererName), "-")),
			fmt.Sprintf("pdf_renderer_path=%s", fallbackSupport(strings.TrimSpace(input.ExportAssets.PDFRendererPath), "-")),
		)
	} else {
		lines = append(lines, "export_assets=unavailable")
	}

	return strings.Join(lines, "\n")
}

func fallbackSupport(value, alt string) string {
	if strings.TrimSpace(value) == "" {
		return alt
	}
	return strings.TrimSpace(value)
}

func pointerString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func optionalInt64String(value *int64) string {
	if value == nil {
		return "-"
	}
	return fmt.Sprintf("%d", *value)
}

func optionalSegmentString(value *sharedtypes.SessionSegmentType) string {
	if value == nil {
		return ""
	}
	return string(*value)
}
