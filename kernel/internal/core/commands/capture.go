package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"crona/kernel/internal/capture"
	"crona/kernel/internal/core"

	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
)

// ParseIssueCapture parses text into proposals without writing anything.
func ParseIssueCapture(ctx context.Context, c *core.Context, input shareddto.IssueCaptureParseRequest) ([]shareddto.IssueCaptureProposal, error) {
	catalog, err := captureCatalog(ctx, c)
	if err != nil {
		return nil, err
	}
	var target *capture.Target
	if input.StreamID != nil && *input.StreamID > 0 {
		target = catalogTarget(catalog, *input.StreamID)
		if target == nil {
			return nil, errors.New("default stream not found")
		}
	} else if name := strings.TrimSpace(input.Target); name != "" {
		resolved, flag := capture.ResolveTarget(name, catalog)
		if resolved == nil {
			if flag == "stream?" {
				return nil, fmt.Errorf("%s has several streams; name one as repo/stream", name)
			}
			return nil, fmt.Errorf("default target: %s", flag)
		}
		target = resolved
	}
	return capture.Parse(input.Text, catalog, target, c.CurrentDate()), nil
}

// ImportIssueCapture creates one issue per proposal. Every proposal is checked before any
// issue is created, so a bad row fails the whole import instead of leaving it half done.
// Proposals may have been edited since parsing, so their flags are not trusted: the
// fields themselves are validated.
func ImportIssueCapture(ctx context.Context, c *core.Context, input shareddto.IssueCaptureImportRequest) (shareddto.IssueCaptureImportResult, error) {
	if len(input.Proposals) == 0 {
		return shareddto.IssueCaptureImportResult{}, errors.New("nothing to import")
	}
	catalog, err := captureCatalog(ctx, c)
	if err != nil {
		return shareddto.IssueCaptureImportResult{}, err
	}
	tags := make([][]string, len(input.Proposals))
	needsInbox := false
	for i, p := range input.Proposals {
		row := fmt.Sprintf("line %d", p.Line)
		if strings.TrimSpace(p.Title) == "" {
			return shareddto.IssueCaptureImportResult{}, fmt.Errorf("%s: title is empty", row)
		}
		if p.EstimateMinutes != nil && *p.EstimateMinutes < 0 {
			return shareddto.IssueCaptureImportResult{}, fmt.Errorf("%s: estimate must be >= 0", row)
		}
		if p.TodoForDate != nil {
			if _, err := time.Parse("2006-01-02", *p.TodoForDate); err != nil {
				return shareddto.IssueCaptureImportResult{}, fmt.Errorf("%s: plan date must be YYYY-MM-DD", row)
			}
		}
		switch {
		case p.StreamID > 0:
			if catalogTarget(catalog, p.StreamID) == nil {
				return shareddto.IssueCaptureImportResult{}, fmt.Errorf("%s: stream not found", row)
			}
		case p.Inbox:
			needsInbox = true
		default:
			return shareddto.IssueCaptureImportResult{}, fmt.Errorf("%s: no stream", row)
		}
		if tags[i], err = sharedtypes.NormalizeTagNames(p.Tags); err != nil {
			return shareddto.IssueCaptureImportResult{}, fmt.Errorf("%s: %w", row, err)
		}
	}
	var inboxStreamID int64
	if needsInbox {
		if inboxStreamID, err = ensureInboxStream(ctx, c, catalog); err != nil {
			return shareddto.IssueCaptureImportResult{}, err
		}
	}
	result := shareddto.IssueCaptureImportResult{Issues: make([]sharedtypes.Issue, 0, len(input.Proposals))}
	for i, p := range input.Proposals {
		streamID := p.StreamID
		if streamID == 0 {
			streamID = inboxStreamID
		}
		created, err := createIssue(ctx, c, issueCreateInput{
			StreamID:        streamID,
			Title:           p.Title,
			Description:     p.Description,
			EstimateMinutes: p.EstimateMinutes,
			TodoForDate:     p.TodoForDate,
		}, false)
		if err != nil {
			return result, fmt.Errorf("line %d: %w", p.Line, err)
		}
		if len(tags[i]) > 0 {
			tagged, err := SetIssueTags(ctx, c, created.ID, tags[i])
			if err != nil {
				return result, fmt.Errorf("line %d: %w", p.Line, err)
			}
			created = *tagged
		}
		result.Issues = append(result.Issues, created)
	}
	return result, nil
}

func captureCatalog(ctx context.Context, c *core.Context) (capture.Catalog, error) {
	repos, err := ListRepos(ctx, c)
	if err != nil {
		return capture.Catalog{}, err
	}
	catalog := capture.Catalog{Repos: repos}
	for _, repo := range repos {
		streams, err := c.Streams.ListByRepo(ctx, repo.ID, c.UserID)
		if err != nil {
			return capture.Catalog{}, err
		}
		catalog.Streams = append(catalog.Streams, streams...)
	}
	return catalog, nil
}

func catalogTarget(catalog capture.Catalog, streamID int64) *capture.Target {
	for _, stream := range catalog.Streams {
		if stream.ID != streamID {
			continue
		}
		for _, repo := range catalog.Repos {
			if repo.ID == stream.RepoID {
				return &capture.Target{RepoID: repo.ID, RepoName: repo.Name, StreamID: stream.ID, StreamName: stream.Name}
			}
		}
	}
	return nil
}

// ensureInboxStream returns the system Inbox repo's Inbox stream, creating either when
// missing. A repo the user named Inbox is used as is, since repo names are unique.
func ensureInboxStream(ctx context.Context, c *core.Context, catalog capture.Catalog) (int64, error) {
	var repoID int64
	for _, repo := range catalog.Repos {
		if repo.System || (repoID == 0 && strings.EqualFold(repo.Name, capture.InboxName)) {
			repoID = repo.ID
		}
	}
	if repoID == 0 {
		repo, err := createRepo(ctx, c, sharedtypes.Repo{Name: capture.InboxName, System: true})
		if err != nil {
			return 0, err
		}
		repoID = repo.ID
	}
	var streamID int64
	for _, stream := range catalog.Streams {
		if stream.RepoID == repoID && (stream.System || (streamID == 0 && strings.EqualFold(stream.Name, capture.InboxName))) {
			streamID = stream.ID
		}
	}
	if streamID > 0 {
		return streamID, nil
	}
	stream, err := createStream(ctx, c, sharedtypes.Stream{
		RepoID: repoID, Name: capture.InboxName, Visibility: sharedtypes.StreamVisibilityPersonal, System: true,
	})
	if err != nil {
		return 0, err
	}
	return stream.ID, nil
}
