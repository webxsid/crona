package controller

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (snapshot Snapshot) Update(msg tea.KeyMsg) (State, *Action, string) {
	return Update(snapshot.Dialog, snapshotUpdateContext(snapshot), snapshot.CurrentDashboardDate, msg)
}

func snapshotUpdateContext(snapshot Snapshot) UpdateContext {
	ctx := UpdateContext{
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
