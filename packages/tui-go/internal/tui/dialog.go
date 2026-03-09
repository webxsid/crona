package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ---------- Dialog open helpers ----------

func (m Model) openCreateScratchpad() Model {
	name := textinput.New()
	name.Placeholder = "My notes"
	name.Focus()
	name.CharLimit = 80
	name.Width = 40

	path := textinput.New()
	path.Placeholder = "notes/[[date]].md"
	path.CharLimit = 120
	path.Width = 40

	m.dialog = "create_scratchpad"
	m.dialogInputs = []textinput.Model{name, path}
	m.dialogFocusIdx = 0
	return m
}

func (m Model) openCreateRepoDialog() Model {
	name := textinput.New()
	name.Placeholder = "Repo name"
	name.Focus()
	name.CharLimit = 80
	name.Width = 40

	m.dialog = "create_repo"
	m.dialogInputs = []textinput.Model{name}
	m.dialogFocusIdx = 0
	return m
}

func (m Model) openCreateStreamDialog(repoID, repoName string) Model {
	name := textinput.New()
	name.Placeholder = "Stream name"
	name.Focus()
	name.CharLimit = 80
	name.Width = 40

	m.dialog = "create_stream"
	m.dialogInputs = []textinput.Model{name}
	m.dialogFocusIdx = 0
	m.dialogRepoID = repoID
	m.dialogRepoName = repoName
	return m
}

func (m Model) openCreateIssueMetaDialog(streamID, streamName, repoName string) Model {
	title := textinput.New()
	title.Placeholder = "Issue title"
	title.Focus()
	title.CharLimit = 120
	title.Width = 52

	estimate := textinput.New()
	estimate.Placeholder = "Estimate minutes (optional)"
	estimate.CharLimit = 6
	estimate.Width = 20

	due := textinput.New()
	due.Placeholder = "Due date YYYY-MM-DD (optional)"
	due.CharLimit = 10
	due.Width = 22

	m.dialog = "create_issue_meta"
	m.dialogInputs = []textinput.Model{title, estimate, due}
	m.dialogFocusIdx = 0
	m.dialogStreamID = streamID
	m.dialogStreamName = streamName
	m.dialogRepoName = repoName
	return m
}

func (m Model) openDatePickerDialog(parentDialog, issueID string, inputIndex int, initial *string) Model {
	selected := m.resolveDialogDate(initial)
	monthStart := time.Date(selected.Year(), selected.Month(), 1, 0, 0, 0, 0, selected.Location())

	m.dialogParent = parentDialog
	m.dialogIssueID = issueID
	m.dialog = "pick_date"
	m.dialogDateCursor = selected.Format("2006-01-02")
	m.dialogDateMonth = monthStart.Format("2006-01-02")
	m.dialogFocusIdx = inputIndex
	return m
}

func (m Model) openCreateIssueDefaultDialog() Model {
	repoFilter := textinput.New()
	repoFilter.Placeholder = "Search repo"
	repoFilter.CharLimit = 80
	repoFilter.Width = 36
	repoFilter.Focus()

	streamFilter := textinput.New()
	streamFilter.Placeholder = "Search stream"
	streamFilter.CharLimit = 80
	streamFilter.Width = 36

	title := textinput.New()
	title.Placeholder = "Issue title"
	title.CharLimit = 120
	title.Width = 52

	estimate := textinput.New()
	estimate.Placeholder = "Estimate minutes (optional)"
	estimate.CharLimit = 6
	estimate.Width = 20

	due := textinput.New()
	due.Placeholder = "Due date YYYY-MM-DD (optional)"
	due.CharLimit = 10
	due.Width = 22

	m.dialog = "create_issue_default"
	m.dialogInputs = []textinput.Model{repoFilter, streamFilter, title, estimate, due}
	m.dialogFocusIdx = 0
	m.dialogRepoIndex = 0
	m.dialogStreamIndex = 0
	return m
}

func (m Model) openConfirmDelete(id string) Model {
	m.dialog = "confirm_delete"
	m.dialogDeleteID = id
	return m
}

func (m Model) openSessionMessageDialog(kind string) Model {
	input := textinput.New()
	if kind == "end_session" {
		input.Placeholder = "End message"
	} else {
		input.Placeholder = "Stash note"
	}
	input.Focus()
	input.CharLimit = 200
	input.Width = 48

	m.dialog = kind
	m.dialogInputs = []textinput.Model{input}
	m.dialogFocusIdx = 0
	return m
}

func (m Model) openIssueSessionTransitionDialog(issueID, status string) Model {
	m.dialog = "issue_session_transition"
	m.dialogIssueID = issueID
	m.dialogIssueStatus = status

	if status == "done" {
		input := textinput.New()
		input.Placeholder = "End message"
		input.Focus()
		input.CharLimit = 200
		input.Width = 48
		m.dialogInputs = []textinput.Model{input}
		m.dialogFocusIdx = 0
	} else {
		m.dialogInputs = nil
		m.dialogFocusIdx = 0
	}

	return m
}

func (m Model) closeDialog() Model {
	m.dialog = ""
	m.dialogInputs = nil
	m.dialogFocusIdx = 0
	m.dialogDeleteID = ""
	m.dialogIssueID = ""
	m.dialogIssueStatus = ""
	m.dialogRepoID = ""
	m.dialogRepoName = ""
	m.dialogStreamID = ""
	m.dialogStreamName = ""
	m.dialogRepoIndex = 0
	m.dialogStreamIndex = 0
	m.dialogParent = ""
	m.dialogDateMonth = ""
	m.dialogDateCursor = ""
	return m
}

// ---------- Dialog keyboard handler ----------

func (m Model) updateDialog(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch m.dialog {
	case "create_repo":
		return m.updateCreateRepoDialog(msg)
	case "create_stream":
		return m.updateCreateStreamDialog(msg)
	case "create_issue_meta":
		return m.updateCreateIssueMetaDialog(msg)
	case "create_issue_default":
		return m.updateCreateIssueDefaultDialog(msg)
	case "pick_date":
		return m.updateDatePickerDialog(msg)
	case "create_scratchpad":
		return m.updateCreateScratchpad(msg)
	case "confirm_delete":
		return m.updateConfirmDelete(msg)
	case "end_session", "stash_session":
		return m.updateSessionMessageDialog(msg)
	case "issue_session_transition":
		return m.updateIssueSessionTransitionDialog(msg)
	}
	return m, nil
}

func (m Model) updateCreateScratchpad(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m.closeDialog(), nil

	case "tab", "shift+tab", "down", "up":
		// cycle focus between inputs
		dir := 1
		if msg.String() == "shift+tab" || msg.String() == "up" {
			dir = -1
		}
		m.dialogFocusIdx = (m.dialogFocusIdx + dir + len(m.dialogInputs)) % len(m.dialogInputs)
		for i := range m.dialogInputs {
			if i == m.dialogFocusIdx {
				m.dialogInputs[i].Focus()
			} else {
				m.dialogInputs[i].Blur()
			}
		}
		return m, nil

	case "enter":
		name := strings.TrimSpace(m.dialogInputs[0].Value())
		path := strings.TrimSpace(m.dialogInputs[1].Value())
		if name == "" || path == "" {
			m.statusMsg = "Name and path are required"
			return m, nil
		}
		m = m.closeDialog()
		return m, cmdCreateScratchpad(m.client, name, path)
	}

	// forward keystrokes to the focused input
	var cmd tea.Cmd
	m.dialogInputs[m.dialogFocusIdx], cmd = m.dialogInputs[m.dialogFocusIdx].Update(msg)
	return m, cmd
}

func (m Model) updateCreateRepoDialog(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m.closeDialog(), nil
	case "enter":
		name := strings.TrimSpace(m.dialogInputs[0].Value())
		if name == "" {
			m.statusMsg = "Repo name is required"
			return m, nil
		}
		m = m.closeDialog()
		return m, cmdCreateRepoOnly(m.client, name)
	}
	var cmd tea.Cmd
	m.dialogInputs[0], cmd = m.dialogInputs[0].Update(msg)
	return m, cmd
}

func (m Model) updateCreateStreamDialog(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m.closeDialog(), nil
	case "enter":
		name := strings.TrimSpace(m.dialogInputs[0].Value())
		if name == "" {
			m.statusMsg = "Stream name is required"
			return m, nil
		}
		m = m.closeDialog()
		return m, cmdCreateStreamOnly(m.client, m.dialogRepoID, name)
	}
	var cmd tea.Cmd
	m.dialogInputs[0], cmd = m.dialogInputs[0].Update(msg)
	return m, cmd
}

func (m Model) updateCreateIssueMetaDialog(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.dialogFocusIdx == 2 && (msg.String() == "f2" || msg.String() == "ctrl+y") {
		return m.openDatePickerDialog("create_issue_meta", "", 2, valueToPointer(m.dialogInputs[2].Value())), nil
	}
	return m.updateMultiInputCreateIssue(msg, 3, func(m Model) (Model, tea.Cmd) {
		title := strings.TrimSpace(m.dialogInputs[0].Value())
		if title == "" {
			m.statusMsg = "Issue title is required"
			return m, nil
		}
		estimate, err := parseEstimateInput(m.dialogInputs[1].Value())
		if err != nil {
			m.statusMsg = err.Error()
			return m, nil
		}
		dueDate, err := parseDueDateInput(m.dialogInputs[2].Value())
		if err != nil {
			m.statusMsg = err.Error()
			return m, nil
		}
		streamID := m.dialogStreamID
		m = m.closeDialog()
		return m, cmdCreateIssueOnly(m.client, streamID, title, estimate, dueDate)
	})
}

func (m Model) updateCreateIssueDefaultDialog(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m.closeDialog(), nil
	case "f2", "ctrl+y":
		if m.dialogFocusIdx == 4 {
			return m.openDatePickerDialog("create_issue_default", "", 4, valueToPointer(m.dialogInputs[4].Value())), nil
		}
	case "tab", "shift+tab", "down", "up":
		if (msg.String() == "down" || msg.String() == "up") && (m.dialogFocusIdx == 0 || m.dialogFocusIdx == 1) {
			if m.dialogFocusIdx == 0 {
				if msg.String() == "down" {
					m.shiftDefaultRepoSelection(1)
				} else {
					m.shiftDefaultRepoSelection(-1)
				}
			} else {
				if msg.String() == "down" {
					m.shiftDefaultStreamSelection(1)
				} else {
					m.shiftDefaultStreamSelection(-1)
				}
			}
			return m, nil
		}
		dir := 1
		if msg.String() == "shift+tab" || msg.String() == "up" {
			dir = -1
		}
		m.dialogFocusIdx = (m.dialogFocusIdx + dir + 5) % 5
		return m.syncDefaultIssueDialogFocus(), nil
	case "left":
		if m.dialogFocusIdx == 0 {
			m.shiftDefaultRepoSelection(-1)
			return m, nil
		}
		if m.dialogFocusIdx == 1 {
			m.shiftDefaultStreamSelection(-1)
			return m, nil
		}
	case "right":
		if m.dialogFocusIdx == 0 {
			m.shiftDefaultRepoSelection(1)
			return m, nil
		}
		if m.dialogFocusIdx == 1 {
			m.shiftDefaultStreamSelection(1)
			return m, nil
		}
	case "enter":
		repoName, streamName := m.defaultIssueDialogNames()
		title := strings.TrimSpace(m.dialogInputs[2].Value())
		if repoName == "" || streamName == "" || title == "" {
			m.statusMsg = "Repo, stream, and issue title are required"
			return m, nil
		}
		estimate, err := parseEstimateInput(m.dialogInputs[3].Value())
		if err != nil {
			m.statusMsg = err.Error()
			return m, nil
		}
		dueDate, err := parseDueDateInput(m.dialogInputs[4].Value())
		if err != nil {
			m.statusMsg = err.Error()
			return m, nil
		}
		m = m.closeDialog()
		return m, cmdCreateIssueWithPath(m.client, repoName, streamName, title, estimate, dueDate)
	}

	var cmd tea.Cmd
	m.dialogInputs[m.dialogFocusIdx], cmd = m.dialogInputs[m.dialogFocusIdx].Update(msg)
	if m.dialogFocusIdx == 0 {
		m.dialogRepoIndex = 0
		m.dialogStreamIndex = 0
	}
	if m.dialogFocusIdx == 1 {
		m.dialogStreamIndex = 0
	}
	return m, cmd
}

func (m Model) updateDatePickerDialog(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m.closeDatePicker(), nil
	case "enter", " ":
		selected := m.dialogDateCursor
		if m.dialogParent == "create_issue_meta" || m.dialogParent == "create_issue_default" {
			if idx := m.dialogFocusIdx; idx >= 0 && idx < len(m.dialogInputs) {
				m.dialogInputs[idx].SetValue(selected)
			}
			return m.closeDatePicker(), nil
		}
		issueID := m.dialogIssueID
		streamID := ""
		if issue := m.activeIssueWithMeta(); issue != nil {
			streamID = issue.StreamID
		}
		m = m.closeDialog()
		return m, cmdSetIssueTodoDate(m.client, issueID, selected, streamID, m.currentDashboardDate())
	case "backspace", "delete", "c":
		if m.dialogParent == "create_issue_meta" || m.dialogParent == "create_issue_default" {
			if idx := m.dialogFocusIdx; idx >= 0 && idx < len(m.dialogInputs) {
				m.dialogInputs[idx].SetValue("")
			}
			return m.closeDatePicker(), nil
		}
		issueID := m.dialogIssueID
		streamID := ""
		if issue := m.activeIssueWithMeta(); issue != nil {
			streamID = issue.StreamID
		}
		m = m.closeDialog()
		return m, cmdSetIssueTodoDate(m.client, issueID, "", streamID, m.currentDashboardDate())
	case "left", "h":
		return m.shiftDatePicker(0, 0, -1), nil
	case "right", "l":
		return m.shiftDatePicker(0, 0, 1), nil
	case "up", "k":
		return m.shiftDatePicker(0, 0, -7), nil
	case "down", "j":
		return m.shiftDatePicker(0, 0, 7), nil
	case ",":
		return m.shiftDatePicker(0, -1, 0), nil
	case ".":
		return m.shiftDatePicker(0, 1, 0), nil
	case "g":
		return m.openDatePickerDialog(m.dialogParent, m.dialogIssueID, m.dialogFocusIdx, valueToPointer(time.Now().Format("2006-01-02"))), nil
	}
	return m, nil
}

func (m Model) updateMultiInputCreateIssue(msg tea.KeyMsg, inputCount int, onSubmit func(Model) (Model, tea.Cmd)) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m.closeDialog(), nil
	case "tab", "shift+tab", "down", "up":
		dir := 1
		if msg.String() == "shift+tab" || msg.String() == "up" {
			dir = -1
		}
		m.dialogFocusIdx = (m.dialogFocusIdx + dir + inputCount) % inputCount
		for i := range m.dialogInputs {
			if i == m.dialogFocusIdx {
				m.dialogInputs[i].Focus()
			} else {
				m.dialogInputs[i].Blur()
			}
		}
		return m, nil
	case "enter":
		return onSubmit(m)
	}

	var cmd tea.Cmd
	m.dialogInputs[m.dialogFocusIdx], cmd = m.dialogInputs[m.dialogFocusIdx].Update(msg)
	return m, cmd
}

func (m Model) updateConfirmDelete(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n", "N":
		return m.closeDialog(), nil
	case "y", "Y", "enter":
		id := m.dialogDeleteID
		m = m.closeDialog()
		return m, cmdDeleteScratchpad(m.client, id)
	}
	return m, nil
}

func (m Model) updateSessionMessageDialog(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m.closeDialog(), nil
	case "enter":
		message := strings.TrimSpace(m.dialogInputs[0].Value())
		kind := m.dialog
		m = m.closeDialog()
		if kind == "end_session" {
			return m, cmdEndFocusSession(m.client, message)
		}
		return m, cmdStashFocusSession(m.client, message)
	}

	var cmd tea.Cmd
	m.dialogInputs[0], cmd = m.dialogInputs[0].Update(msg)
	return m, cmd
}

func (m Model) updateIssueSessionTransitionDialog(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n", "N":
		return m.closeDialog(), nil
	case "y", "Y", "enter":
		issueID := m.dialogIssueID
		status := m.dialogIssueStatus
		message := ""
		if status == "done" && len(m.dialogInputs) > 0 {
			message = strings.TrimSpace(m.dialogInputs[0].Value())
		}
		streamID := ""
		if issue := m.activeIssueWithMeta(); issue != nil {
			streamID = issue.StreamID
		}
		m = m.closeDialog()
		return m, cmdChangeIssueStatusAndEndSession(m.client, issueID, status, streamID, m.currentDashboardDate(), message)
	}

	if m.dialogIssueStatus == "done" && len(m.dialogInputs) > 0 {
		var cmd tea.Cmd
		m.dialogInputs[0], cmd = m.dialogInputs[0].Update(msg)
		return m, cmd
	}
	return m, nil
}

// ---------- Open in $EDITOR ----------

// openEditor suspends the TUI and opens the file at path in $EDITOR.
// When the editor exits Bubbletea resumes and we reload scratchpads.
func openEditor(filePath string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
	}
	c := exec.Command(editor, filePath)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return errMsg{err}
		}
		return editorDoneMsg{}
	})
}

type editorDoneMsg struct{}

// ---------- Dialog rendering ----------

func (m Model) renderDialog() string {
	switch m.dialog {
	case "create_repo":
		return m.renderCreateRepoDialog()
	case "create_stream":
		return m.renderCreateStreamDialog()
	case "create_issue_meta":
		return m.renderCreateIssueMetaDialog()
	case "create_issue_default":
		return m.renderCreateIssueDefaultDialog()
	case "pick_date":
		return m.renderDatePickerDialog()
	case "create_scratchpad":
		return m.renderCreateScratchpadDialog()
	case "confirm_delete":
		return m.renderConfirmDeleteDialog()
	case "end_session", "stash_session":
		return m.renderSessionMessageDialog()
	case "issue_session_transition":
		return m.renderIssueSessionTransitionDialog()
	}
	return ""
}

func (m Model) renderCreateScratchpadDialog() string {
	title := stylePaneTitle.Render("New Scratchpad")
	rows := []string{
		title,
		"",
		styleDim.Render("Name"),
		m.dialogInputs[0].View(),
		"",
		styleDim.Render("Path  (supports [[date]], [[timestamp]])"),
		m.dialogInputs[1].View(),
		"",
		styleDim.Render("[tab] next field   [enter] create   [esc] cancel"),
	}

	content := strings.Join(rows, "\n")
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorCyan).
		Padding(1, 3).
		Width(min(m.width-8, 54)).
		Render(content)
}

func (m Model) renderCreateRepoDialog() string {
	rows := []string{
		stylePaneTitle.Render("New Repo"),
		"",
		styleDim.Render("Name"),
		m.dialogInputs[0].View(),
		"",
		styleDim.Render("[enter] create   [esc] cancel"),
	}
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorCyan).
		Padding(1, 3).
		Width(min(m.width-8, 52)).
		Render(strings.Join(rows, "\n"))
}

func (m Model) renderCreateStreamDialog() string {
	rows := []string{
		stylePaneTitle.Render("New Stream"),
		"",
		styleDim.Render("Repo"),
		styleHeader.Render(m.dialogRepoName),
		"",
		styleDim.Render("Name"),
		m.dialogInputs[0].View(),
		"",
		styleDim.Render("[enter] create   [esc] cancel"),
	}
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorCyan).
		Padding(1, 3).
		Width(min(m.width-8, 56)).
		Render(strings.Join(rows, "\n"))
}

func (m Model) renderCreateIssueMetaDialog() string {
	rows := []string{
		stylePaneTitle.Render("New Issue"),
		"",
		styleDim.Render("Repo / Stream"),
		styleHeader.Render(m.dialogRepoName + " / " + m.dialogStreamName),
		"",
		styleDim.Render("Title"),
		m.dialogInputs[0].View(),
		"",
		styleDim.Render("Estimate"),
		m.dialogInputs[1].View(),
		"",
		styleDim.Render("Due"),
		m.dialogInputs[2].View(),
		"",
		styleDim.Render("[f2] calendar   [tab] next   [enter] create   [esc] cancel"),
	}
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorCyan).
		Padding(1, 3).
		Width(min(m.width-8, 68)).
		Render(strings.Join(rows, "\n"))
}

func (m Model) renderCreateIssueDefaultDialog() string {
	repoLabel, streamLabel := m.defaultIssueDialogLabels()
	rows := []string{
		stylePaneTitle.Render("New Issue"),
		"",
		styleDim.Render("Repo"),
		m.dialogInputs[0].View(),
		"",
		m.renderDialogSelector(repoLabel, m.dialogFocusIdx == 0),
		"",
		styleDim.Render("Stream"),
		m.dialogInputs[1].View(),
		"",
		m.renderDialogSelector(streamLabel, m.dialogFocusIdx == 1),
		"",
		styleDim.Render("Title"),
		m.dialogInputs[2].View(),
		"",
		styleDim.Render("Estimate"),
		m.dialogInputs[3].View(),
		"",
		styleDim.Render("Due"),
		m.dialogInputs[4].View(),
		"",
		styleDim.Render("[type] filter   [up/down] choose   [f2] calendar   [tab] next   [enter] create"),
	}
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorCyan).
		Padding(1, 3).
		Width(min(m.width-8, 72)).
		Render(strings.Join(rows, "\n"))
}

func (m Model) renderConfirmDeleteDialog() string {
	name := ""
	for _, s := range m.scratchpads {
		if s.ID == m.dialogDeleteID {
			name = s.Name
			break
		}
	}
	content := fmt.Sprintf(
		"%s\n\nDelete %s?\n\n%s",
		stylePaneTitle.Render("Confirm Delete"),
		styleError.Render(name),
		styleDim.Render("[y] yes   [n/esc] cancel"),
	)
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorRed).
		Padding(1, 3).
		Width(min(m.width-8, 44)).
		Render(content)
}

func (m Model) renderSessionMessageDialog() string {
	title := "End Session"
	hint := "[enter] confirm   [esc] cancel"
	if m.dialog == "stash_session" {
		title = "Stash Session"
	}

	rows := []string{
		stylePaneTitle.Render(title),
		"",
		styleDim.Render(title + " message"),
		m.dialogInputs[0].View(),
		"",
		styleDim.Render(hint),
	}

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorCyan).
		Padding(1, 3).
		Width(min(m.width-8, 64)).
		Render(strings.Join(rows, "\n"))
}

func (m Model) renderIssueSessionTransitionDialog() string {
	title := "End Session?"
	body := "Mark this issue and end the active session?"
	hint := "[y/enter] confirm   [n/esc] cancel"
	border := colorYellow

	if m.dialogIssueStatus == "done" {
		title = "Complete Issue"
		body = "Mark the issue done and end the active session."
		border = colorGreen
	} else if m.dialogIssueStatus == "abandoned" {
		title = "Abandon Issue"
		body = "Abandon the issue and end the active session."
		border = colorRed
	}

	rows := []string{
		stylePaneTitle.Render(title),
		"",
		body,
	}
	if m.dialogIssueStatus == "done" && len(m.dialogInputs) > 0 {
		rows = append(rows,
			"",
			styleDim.Render("End message"),
			m.dialogInputs[0].View(),
		)
	}
	rows = append(rows, "", styleDim.Render(hint))

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(1, 3).
		Width(min(m.width-8, 68)).
		Render(strings.Join(rows, "\n"))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func parseEstimateInput(raw string) (*int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return nil, fmt.Errorf("Estimate must be a non-negative integer")
	}
	return &value, nil
}

func parseDueDateInput(raw string) (*string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if _, err := time.Parse("2006-01-02", raw); err != nil {
		return nil, fmt.Errorf("Due date must be YYYY-MM-DD")
	}
	return &raw, nil
}

func (m Model) renderDatePickerDialog() string {
	selected := m.dialogDate()
	monthStart := m.dialogMonth()
	grid := m.renderCalendarGrid(monthStart, selected)

	title := "Pick Due Date"
	if m.dialogParent == "create_issue_meta" || m.dialogParent == "create_issue_default" {
		title = "Pick Due Date For New Issue"
	}

	rows := []string{
		stylePaneTitle.Render(title),
		"",
		styleHeader.Render(selected.Format("Mon, 02 Jan 2006")),
		styleDim.Render(monthStart.Format("January 2006")),
		"",
		grid,
		"",
		styleDim.Render("[h/j/k/l] move   [,/.] month   [enter] choose   [c] clear   [esc] back"),
	}

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorCyan).
		Padding(1, 3).
		Width(min(m.width-8, 46)).
		Render(strings.Join(rows, "\n"))
}

type selectorOption struct {
	id    string
	label string
}

func (m Model) defaultRepoOptions() []selectorOption {
	query := strings.TrimSpace(strings.ToLower(m.dialogInputs[0].Value()))
	options := make([]selectorOption, 0, len(m.repos)+1)
	for _, repo := range m.repos {
		if query != "" && !strings.Contains(strings.ToLower(repo.Name), query) {
			continue
		}
		options = append(options, selectorOption{id: repo.ID, label: repo.Name})
	}
	if raw := strings.TrimSpace(m.dialogInputs[0].Value()); raw != "" {
		options = append(options, selectorOption{id: "__new__", label: "Create New Repo: " + raw})
	}
	return options
}

func (m Model) defaultStreamOptions() []selectorOption {
	query := strings.TrimSpace(strings.ToLower(m.dialogInputs[1].Value()))
	if m.defaultRepoIsNew() {
		if raw := strings.TrimSpace(m.dialogInputs[1].Value()); raw != "" {
			return []selectorOption{{id: "__new__", label: "Create New Stream: " + raw}}
		}
		return []selectorOption{}
	}
	repoOptions := m.defaultRepoOptions()
	repoOpt := repoOptions[minInt(m.dialogRepoIndex, len(repoOptions)-1)]

	seen := map[string]bool{}
	options := []selectorOption{}
	for _, issue := range m.allIssues {
		if issue.RepoID != repoOpt.id || seen[issue.StreamID] {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(issue.StreamName), query) {
			continue
		}
		seen[issue.StreamID] = true
		options = append(options, selectorOption{id: issue.StreamID, label: issue.StreamName})
	}
	if m.context != nil && m.context.RepoID != nil && *m.context.RepoID == repoOpt.id {
		for _, stream := range m.streams {
			if !seen[stream.ID] {
				if query != "" && !strings.Contains(strings.ToLower(stream.Name), query) {
					continue
				}
				seen[stream.ID] = true
				options = append(options, selectorOption{id: stream.ID, label: stream.Name})
			}
		}
	}
	if raw := strings.TrimSpace(m.dialogInputs[1].Value()); raw != "" {
		options = append(options, selectorOption{id: "__new__", label: "Create New Stream: " + raw})
	}
	return options
}

func (m Model) defaultRepoIsNew() bool {
	options := m.defaultRepoOptions()
	if len(options) == 0 {
		return true
	}
	return options[minInt(m.dialogRepoIndex, len(options)-1)].id == "__new__"
}

func (m Model) defaultStreamIsNew() bool {
	options := m.defaultStreamOptions()
	if len(options) == 0 {
		return true
	}
	return options[minInt(m.dialogStreamIndex, len(options)-1)].id == "__new__"
}

func (m *Model) shiftDefaultRepoSelection(dir int) {
	options := m.defaultRepoOptions()
	if len(options) == 0 {
		return
	}
	m.dialogRepoIndex = (m.dialogRepoIndex + dir + len(options)) % len(options)
	m.dialogStreamIndex = 0
}

func (m *Model) shiftDefaultStreamSelection(dir int) {
	options := m.defaultStreamOptions()
	if len(options) == 0 {
		return
	}
	m.dialogStreamIndex = (m.dialogStreamIndex + dir + len(options)) % len(options)
}

func (m Model) defaultIssueDialogFieldCount() int {
	return 4
}

func (m Model) syncDefaultIssueDialogFocus() Model {
	for i := range m.dialogInputs {
		m.dialogInputs[i].Blur()
	}
	if m.dialogFocusIdx >= 0 && m.dialogFocusIdx < len(m.dialogInputs) {
		m.dialogInputs[m.dialogFocusIdx].Focus()
	}
	return m
}

func (m Model) defaultIssueDialogLabels() (string, string) {
	repoOptions := m.defaultRepoOptions()
	streamOptions := m.defaultStreamOptions()
	if len(repoOptions) == 0 {
		return "Type to search or create", "Select a repo first"
	}
	if len(streamOptions) == 0 {
		return repoOptions[minInt(m.dialogRepoIndex, len(repoOptions)-1)].label, "Type to search or create"
	}
	repo := repoOptions[minInt(m.dialogRepoIndex, len(repoOptions)-1)].label
	stream := streamOptions[minInt(m.dialogStreamIndex, len(streamOptions)-1)].label
	return repo, stream
}

func (m Model) defaultIssueDialogNames() (string, string) {
	repoOptions := m.defaultRepoOptions()
	streamOptions := m.defaultStreamOptions()
	if len(repoOptions) == 0 || len(streamOptions) == 0 {
		return "", ""
	}
	repo := repoOptions[minInt(m.dialogRepoIndex, len(repoOptions)-1)]
	stream := streamOptions[minInt(m.dialogStreamIndex, len(streamOptions)-1)]

	repoName := repo.label
	if repo.id == "__new__" {
		repoName = strings.TrimSpace(m.dialogInputs[0].Value())
	}
	streamName := stream.label
	if stream.id == "__new__" {
		streamName = strings.TrimSpace(m.dialogInputs[1].Value())
	}
	return repoName, streamName
}

func (m Model) renderDialogSelector(label string, active bool) string {
	style := styleNormal
	if active {
		style = styleCursor
	}
	return style.Render("[ " + label + " ]")
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func valueToPointer(raw string) *string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	return &raw
}

func (m Model) resolveDialogDate(initial *string) time.Time {
	if initial != nil {
		if parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*initial)); err == nil {
			return parsed
		}
	}
	if parsed, err := time.Parse("2006-01-02", m.currentDashboardDate()); err == nil {
		return parsed
	}
	return time.Now()
}

func (m Model) dialogDate() time.Time {
	if parsed, err := time.Parse("2006-01-02", m.dialogDateCursor); err == nil {
		return parsed
	}
	return m.resolveDialogDate(nil)
}

func (m Model) dialogMonth() time.Time {
	if parsed, err := time.Parse("2006-01-02", m.dialogDateMonth); err == nil {
		return parsed
	}
	date := m.dialogDate()
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
}

func (m Model) closeDatePicker() Model {
	parent := m.dialogParent
	dateField := m.dialogFocusIdx
	m.dialog = parent
	m.dialogParent = ""
	m.dialogDateMonth = ""
	m.dialogDateCursor = ""
	if parent == "create_issue_meta" || parent == "create_issue_default" {
		m.dialogFocusIdx = dateField
		return m.syncDefaultIssueDialogFocusForParent()
	}
	return m
}

func (m Model) syncDefaultIssueDialogFocusForParent() Model {
	if m.dialog == "create_issue_default" {
		return m.syncDefaultIssueDialogFocus()
	}
	for i := range m.dialogInputs {
		if i == m.dialogFocusIdx {
			m.dialogInputs[i].Focus()
		} else {
			m.dialogInputs[i].Blur()
		}
	}
	return m
}

func (m Model) shiftDatePicker(years, months, days int) Model {
	selected := m.dialogDate().AddDate(years, months, days)
	monthStart := time.Date(selected.Year(), selected.Month(), 1, 0, 0, 0, 0, selected.Location())
	m.dialogDateCursor = selected.Format("2006-01-02")
	m.dialogDateMonth = monthStart.Format("2006-01-02")
	return m
}

func (m Model) renderCalendarGrid(monthStart, selected time.Time) string {
	headers := []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
	lines := []string{strings.Join(headers, "  ")}

	offset := (int(monthStart.Weekday()) + 6) % 7
	gridStart := monthStart.AddDate(0, 0, -offset)

	for week := 0; week < 6; week++ {
		cells := make([]string, 0, 7)
		for day := 0; day < 7; day++ {
			current := gridStart.AddDate(0, 0, week*7+day)
			label := fmt.Sprintf("%2d", current.Day())
			cell := " " + label + " "
			style := styleNormal
			if current.Month() != monthStart.Month() {
				style = styleDim
			}
			if sameDay(current, selected) {
				cell = styleCursor.Render(cell)
			} else {
				cell = style.Render(cell)
			}
			cells = append(cells, cell)
		}
		lines = append(lines, strings.Join(cells, " "))
	}

	return strings.Join(lines, "\n")
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
