// Package capture parses `-` lines into issue proposals for bulk import. It is pure:
// the caller supplies the repos and streams to match against and today's date.
package capture

import (
	"strings"

	shareddto "crona/shared/dto"
	sharedtypes "crona/shared/types"
)

// InboxName names the built-in Inbox repo and its single stream.
const InboxName = "Inbox"

type Catalog struct {
	Repos   []sharedtypes.Repo
	Streams []sharedtypes.Stream
}

// Target is where a line goes without a `#` annotation.
type Target struct {
	RepoID     int64
	RepoName   string
	StreamID   int64
	StreamName string
}

// Parse turns every top-level `- ` line of text into a proposal. Lines indented under it
// become its description; lines in fenced code blocks and all other lines are ignored.
// A nil target sends untargeted lines to the Inbox stream.
func Parse(text string, catalog Catalog, target *Target, today string) []shareddto.IssueCaptureProposal {
	var out []shareddto.IssueCaptureProposal
	var description []string
	current := -1
	flush := func() {
		if current >= 0 && len(description) > 0 {
			joined := strings.Join(description, "\n")
			out[current].Description = &joined
		}
		description = nil
	}
	fenced := false
	for i, raw := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(raw)
		if strings.HasPrefix(trimmed, "```") {
			fenced = !fenced
			continue
		}
		if fenced {
			continue
		}
		indented := strings.HasPrefix(raw, " ") || strings.HasPrefix(raw, "\t")
		switch {
		case trimmed == "":
		case indented && current >= 0:
			description = append(description, trimmed)
		case !indented && strings.HasPrefix(raw, "- "):
			flush()
			out = append(out, parseLine(i+1, raw[2:], catalog, target, today))
			current = len(out) - 1
		default:
			flush()
			current = -1
		}
	}
	flush()
	return out
}

func parseLine(number int, text string, catalog Catalog, target *Target, today string) shareddto.IssueCaptureProposal {
	p := shareddto.IssueCaptureProposal{Line: number}
	var title []string
	targeted := false
	for _, word := range strings.Fields(text) {
		switch {
		case len(word) > 1 && word[0] == '#' && !targeted:
			targeted = true
			resolveTarget(&p, word[1:], catalog)
		case len(word) > 1 && word[0] == '~' && p.EstimateMinutes == nil:
			if minutes, ok := parseDuration(word[1:]); ok {
				p.EstimateMinutes = &minutes
				continue
			}
			title = append(title, word)
		case len(word) > 1 && word[0] == '@' && p.TodoForDate == nil:
			if date, ok := parseDateWord(word[1:], today); ok {
				p.TodoForDate = &date
				continue
			}
			p.Flags = append(p.Flags, "bad date "+word)
			title = append(title, word)
		case len(word) > 1 && word[0] == '+':
			name, err := sharedtypes.NormalizeTagName(word[1:])
			if err != nil {
				p.Flags = append(p.Flags, "bad tag "+word)
				title = append(title, word)
				continue
			}
			if !contains(p.Tags, name) {
				p.Tags = append(p.Tags, name)
			}
		default:
			title = append(title, word)
		}
	}
	title = trailingPhrases(&p, title, today)
	p.Title = strings.Join(title, " ")
	if p.Title == "" {
		p.Flags = append(p.Flags, "no title")
	}
	if !targeted {
		applyDefault(&p, target)
	}
	return p
}

// trailingPhrases takes natural date and estimate phrases off the end of the title, one
// per field and only for fields no annotation already set.
func trailingPhrases(p *shareddto.IssueCaptureProposal, title []string, today string) []string {
	for range 2 {
		if p.TodoForDate == nil {
			if date, n := trailingDate(title, today); n > 0 && n < len(title) {
				p.TodoForDate = &date
				title = title[:len(title)-n]
				continue
			}
		}
		if p.EstimateMinutes == nil {
			if minutes, n := trailingEstimate(title); n > 0 && n < len(title) {
				p.EstimateMinutes = &minutes
				title = title[:len(title)-n]
				continue
			}
		}
		break
	}
	return title
}

func applyDefault(p *shareddto.IssueCaptureProposal, target *Target) {
	if target == nil {
		p.Inbox = true
		p.RepoName, p.StreamName = InboxName, InboxName
		return
	}
	p.RepoID, p.RepoName = target.RepoID, target.RepoName
	p.StreamID, p.StreamName = target.StreamID, target.StreamName
}

// resolveTarget matches `#repo/stream` or `#name` by case-insensitive prefix, preferring
// exact names. Anything it cannot settle on one stream is flagged.
func resolveTarget(p *shareddto.IssueCaptureProposal, name string, catalog Catalog) {
	if repoName, streamName, ok := strings.Cut(name, "/"); ok {
		repos := matchRepos(catalog.Repos, repoName)
		if len(repos) != 1 {
			p.Flags = append(p.Flags, targetFlag(len(repos), "#"+name))
			return
		}
		setRepo(p, repos[0])
		streams := matchStreams(streamsOf(catalog.Streams, repos[0].ID), streamName)
		if len(streams) != 1 {
			p.Flags = append(p.Flags, targetFlag(len(streams), "#"+name))
			return
		}
		setStream(p, streams[0])
		return
	}
	repos := matchRepos(catalog.Repos, name)
	streams := matchStreams(catalog.Streams, name)
	switch {
	case len(repos) == 0 && len(streams) == 1:
		setStream(p, streams[0])
		for _, repo := range catalog.Repos {
			if repo.ID == streams[0].RepoID {
				setRepo(p, repo)
			}
		}
	case len(repos) == 1 && len(streams) == 0:
		setRepo(p, repos[0])
		if own := streamsOf(catalog.Streams, repos[0].ID); len(own) == 1 {
			setStream(p, own[0])
		} else {
			p.Flags = append(p.Flags, "stream?")
		}
	case len(repos) == 1 && len(streams) == 1 && streams[0].RepoID == repos[0].ID:
		setRepo(p, repos[0])
		setStream(p, streams[0])
	default:
		p.Flags = append(p.Flags, targetFlag(len(repos)+len(streams), "#"+name))
	}
}

func targetFlag(matches int, token string) string {
	if matches == 0 {
		return "unknown " + token
	}
	return "ambiguous " + token
}

func setRepo(p *shareddto.IssueCaptureProposal, repo sharedtypes.Repo) {
	p.RepoID, p.RepoName = repo.ID, repo.Name
}

func setStream(p *shareddto.IssueCaptureProposal, stream sharedtypes.Stream) {
	p.StreamID, p.StreamName = stream.ID, stream.Name
}

func matchRepos(repos []sharedtypes.Repo, name string) []sharedtypes.Repo {
	return matchNames(repos, name, func(r sharedtypes.Repo) string { return r.Name })
}

func matchStreams(streams []sharedtypes.Stream, name string) []sharedtypes.Stream {
	return matchNames(streams, name, func(s sharedtypes.Stream) string { return s.Name })
}

func matchNames[T any](items []T, name string, nameOf func(T) string) []T {
	name = strings.ToLower(name)
	if name == "" {
		return nil
	}
	var exact, prefix []T
	for _, item := range items {
		switch candidate := strings.ToLower(nameOf(item)); {
		case candidate == name:
			exact = append(exact, item)
		case strings.HasPrefix(candidate, name):
			prefix = append(prefix, item)
		}
	}
	if len(exact) > 0 {
		return exact
	}
	return prefix
}

func streamsOf(streams []sharedtypes.Stream, repoID int64) []sharedtypes.Stream {
	var out []sharedtypes.Stream
	for _, stream := range streams {
		if stream.RepoID == repoID {
			out = append(out, stream)
		}
	}
	return out
}

func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
