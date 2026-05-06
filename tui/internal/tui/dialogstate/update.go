package dialogstate

import (
	dialogpkg "crona/tui/internal/tui/dialogs"

	tea "github.com/charmbracelet/bubbletea"
)

func Update(snapshot Snapshot, msg tea.KeyMsg) (dialogpkg.State, *dialogpkg.Action, string) {
	return dialogpkg.Update(snapshot.Dialog, UpdateContext(snapshot), snapshot.CurrentDashboardDate, msg)
}

func UpdateContext(snapshot Snapshot) dialogpkg.UpdateContext {
	ctx := dialogpkg.UpdateContext{
		Repos:    snapshot.Repos,
		Streams:  snapshot.Streams,
		AllIssues: snapshot.AllIssues,
		Context:  snapshot.Context,
		Stashes:  snapshot.Stashes,
	}
	if snapshot.HasSelectedIssue {
		ctx.SelectedIssueID = snapshot.SelectedIssueID
		ctx.SelectedStreamID = snapshot.SelectedStreamID
		ctx.HasSelectedIssue = true
	}
	if snapshot.HasActiveIssue {
		ctx.ActiveIssueStream = snapshot.ActiveIssueStream
		ctx.HasActiveIssue = true
	}
	return ctx
}
