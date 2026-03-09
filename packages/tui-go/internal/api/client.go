package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// Client is an authenticated HTTP client for the Crona kernel.
type Client struct {
	baseURL    string
	token      string
	scratchDir string
	http       *http.Client
}

func NewClient(baseURL, token, scratchDir string) *Client {
	return &Client{
		baseURL:    baseURL,
		token:      token,
		scratchDir: scratchDir,
		http:       &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) get(path string, out any) error {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GET %s → %d: %s", path, resp.StatusCode, body)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) post(path string, body, out any) error {
	return c.sendJSON(http.MethodPost, path, body, out)
}

func (c *Client) put(path string, body, out any) error {
	return c.sendJSON(http.MethodPut, path, body, out)
}

func (c *Client) delete(path string) error {
	req, err := http.NewRequest(http.MethodDelete, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("DELETE %s → %d", path, resp.StatusCode)
	}
	return nil
}

func (c *Client) sendJSON(method, path string, body, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(method, c.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s %s → %d: %s", method, path, resp.StatusCode, body)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

// ---------- Repos ----------

func (c *Client) ListRepos() ([]Repo, error) {
	var out []Repo
	return out, c.get("/repos", &out)
}

func (c *Client) CreateRepo(name string) (*Repo, error) {
	var out Repo
	return &out, c.post("/commands/repo", map[string]string{"name": name}, &out)
}

// ---------- Streams ----------

func (c *Client) ListStreams(repoID string) ([]Stream, error) {
	var out []Stream
	return out, c.get("/streams?repoId="+repoID, &out)
}

func (c *Client) CreateStream(repoID, name string) (*Stream, error) {
	var out Stream
	return &out, c.post("/stream", map[string]string{"repoId": repoID, "name": name}, &out)
}

// ---------- Issues ----------

func (c *Client) ListIssues(streamID string) ([]Issue, error) {
	var out []Issue
	return out, c.get("/issues?streamId="+streamID, &out)
}

func (c *Client) ListAllIssues() ([]IssueWithMeta, error) {
	var out []IssueWithMeta
	return out, c.get("/issues/all", &out)
}

func (c *Client) CreateIssue(streamID, title string, estimateMinutes *int, todoForDate *string) (*Issue, error) {
	body := map[string]any{"streamId": streamID, "title": title}
	if estimateMinutes != nil {
		body["estimateMinutes"] = *estimateMinutes
	}
	if todoForDate != nil && strings.TrimSpace(*todoForDate) != "" {
		body["todoForDate"] = *todoForDate
	}
	var out Issue
	return &out, c.post("/issue", body, &out)
}

func (c *Client) ListSessionsByIssue(issueID string) ([]Session, error) {
	var out []Session
	return out, c.get("/sessions?issueId="+issueID, &out)
}

func (c *Client) GetDailySummary(date string) (*DailyIssueSummary, error) {
	var out DailyIssueSummary
	path := "/issues/summary/daily"
	if strings.TrimSpace(date) != "" {
		path += "?date=" + date
	}
	return &out, c.get(path, &out)
}

func (c *Client) ChangeIssueStatus(issueID, status string) error {
	return c.put("/issue/"+issueID+"/status", map[string]string{"status": status}, nil)
}

func (c *Client) MarkIssueTodoForToday(issueID string) error {
	return c.SetIssueTodoDate(issueID, "")
}

func (c *Client) SetIssueTodoDate(issueID, date string) error {
	body := map[string]any{}
	if strings.TrimSpace(date) != "" {
		body["date"] = strings.TrimSpace(date)
	}
	return c.put("/issue/"+issueID+"/todo", body, nil)
}

func (c *Client) ClearIssueTodo(issueID string) error {
	return c.put("/issue/"+issueID+"/todo/clear", map[string]any{}, nil)
}

// ---------- Context ----------

func (c *Client) GetContext() (*ActiveContext, error) {
	var out ActiveContext
	return &out, c.get("/context", &out)
}

func (c *Client) SwitchRepo(repoID string) error {
	return c.put("/context/repo", map[string]string{"repoId": repoID}, nil)
}

func (c *Client) SwitchStream(streamID string) error {
	return c.put("/context/stream", map[string]string{"streamId": streamID}, nil)
}

func (c *Client) SwitchIssue(issueID string) error {
	return c.put("/context/issue", map[string]string{"issueId": issueID}, nil)
}

func (c *Client) SetFullContext(repoID, streamID, issueID string) error {
	return c.put("/context", map[string]string{
		"repoId":   repoID,
		"streamId": streamID,
		"issueId":  issueID,
	}, nil)
}

// ---------- Timer ----------

func (c *Client) GetTimerState() (*TimerState, error) {
	var out TimerState
	return &out, c.get("/timer/state", &out)
}

func (c *Client) GetHealth() (*Health, error) {
	var out Health
	return &out, c.get("/health", &out)
}

func (c *Client) ShutdownKernel() error {
	return c.post("/kernel/shutdown", map[string]any{}, nil)
}

func (c *Client) StartTimer(issueID string) error {
	path := "/timer/start"
	if issueID != "" {
		path += "?issueId=" + issueID
	}
	return c.post(path, nil, nil)
}

func (c *Client) PauseTimer() error {
	return c.post("/timer/pause", nil, nil)
}

func (c *Client) ResumeTimer() error {
	return c.post("/timer/resume", nil, nil)
}

func (c *Client) EndTimer(commitMessage string) error {
	body := map[string]string{}
	if commitMessage != "" {
		body["commitMessage"] = commitMessage
	}
	return c.post("/timer/end", body, nil)
}

func (c *Client) StashPush(note string) error {
	body := map[string]string{}
	if note != "" {
		body["stashNote"] = note
	}
	return c.post("/stash", body, nil)
}

// ---------- Scratchpads ----------

func (c *Client) ListScratchpads() ([]ScratchPad, error) {
	var out []ScratchPad
	return out, c.get("/scratchpads", &out)
}

func (c *Client) RegisterScratchpad(id, name, path string) error {
	body := map[string]any{
		"id":           id,
		"name":         name,
		"path":         path,
		"pinned":       false,
		"lastOpenedAt": time.Now().UTC().Format(time.RFC3339),
	}
	return c.post("/scratchpads/register", body, nil)
}

func (c *Client) ReadScratchpad(id string) (string, string, error) {
	var out struct {
		OK      bool        `json:"ok"`
		Meta    *ScratchPad `json:"meta"`
		Content string      `json:"content"`
	}
	if err := c.get("/scratchpads/read/"+id, &out); err != nil {
		return "", "", err
	}
	path := ""
	if out.Meta != nil {
		relativePath := out.Meta.Path
		if !strings.HasSuffix(relativePath, ".md") {
			relativePath += ".md"
		}
		if c.scratchDir != "" {
			path = filepath.Join(c.scratchDir, relativePath)
		} else {
			path = relativePath
		}
	}
	return path, out.Content, nil
}

func (c *Client) DeleteScratchpad(id string) error {
	return c.delete("/scratchpads/" + id)
}

// ---------- Ops ----------

func (c *Client) ListOps(limit int) ([]Op, error) {
	var out []Op
	if limit <= 0 {
		limit = 50
	}
	return out, c.get(fmt.Sprintf("/ops/latest?limit=%d", limit), &out)
}
