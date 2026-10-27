package dto

import "crona/shared/types"

// IssueCaptureParseRequest parses `-` lines into issue proposals. StreamID, or Target as
// a `repo/stream` or single name matched like a `#` annotation, is the caller's default
// target; without either, lines with no target go to the Inbox stream.
type IssueCaptureParseRequest struct {
	Text     string `json:"text"`
	StreamID *int64 `json:"streamId,omitempty"`
	Target   string `json:"target,omitempty"`
}

// IssueCaptureProposal is one parsed line. A proposal with flags cannot be imported
// until its fields are fixed. Inbox marks the built-in Inbox stream, which is created
// on import when it does not exist yet (StreamID is then 0).
type IssueCaptureProposal struct {
	Line            int      `json:"line"`
	Title           string   `json:"title"`
	Description     *string  `json:"description,omitempty"`
	EstimateMinutes *int     `json:"estimateMinutes,omitempty"`
	TodoForDate     *string  `json:"todoForDate,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	RepoID          int64    `json:"repoId,omitempty"`
	RepoName        string   `json:"repoName,omitempty"`
	StreamID        int64    `json:"streamId,omitempty"`
	StreamName      string   `json:"streamName,omitempty"`
	Inbox           bool     `json:"inbox,omitempty"`
	Flags           []string `json:"flags,omitempty"`
}

type IssueCaptureImportRequest struct {
	Proposals []IssueCaptureProposal `json:"proposals"`
}

type IssueCaptureImportResult struct {
	Issues []types.Issue `json:"issues"`
}
