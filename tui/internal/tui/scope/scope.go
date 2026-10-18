// Package scope is the Issues tab's browse scope: All, a repo, or one of its streams.
package scope

import "crona/tui/internal/api"

// Scope is a repo and optionally one of its streams; the zero value is All.
type Scope struct{ RepoID, StreamID int64 }

func (s Scope) IsAll() bool { return s.RepoID == 0 }

// Contains reports whether an item in repoID and streamID falls within the scope.
func (s Scope) Contains(repoID, streamID int64) bool {
	switch {
	case s.IsAll():
		return true
	case s.StreamID != 0:
		return s.StreamID == streamID
	default:
		return s.RepoID == repoID
	}
}

// Node is one row of the scope tree.
type Node struct {
	Scope Scope
	Name  string
	Depth int
}

// Nodes is All, then each repo followed by its streams, in the order given.
func Nodes(repos []api.Repo, streams []api.Stream) []Node {
	nodes := []Node{{Name: "All"}}
	for _, repo := range repos {
		nodes = append(nodes, Node{Scope: Scope{RepoID: repo.ID}, Name: repo.Name})
		for _, stream := range streams {
			if stream.RepoID == repo.ID {
				nodes = append(nodes, Node{Scope: Scope{RepoID: repo.ID, StreamID: stream.ID}, Name: stream.Name, Depth: 1})
			}
		}
	}
	return nodes
}

// IndexOf is s's node, else its repo's node, else All.
func IndexOf(nodes []Node, s Scope) int {
	repo := 0
	for i, node := range nodes {
		if node.Scope == s {
			return i
		}
		if repo == 0 && node.Depth == 0 && !node.Scope.IsAll() && node.Scope.RepoID == s.RepoID {
			repo = i
		}
	}
	return repo
}

// FromContext is the checked-out repo and stream as a scope.
func FromContext(ctx *api.ActiveContext) Scope {
	if ctx == nil || ctx.RepoID == nil {
		return Scope{}
	}
	s := Scope{RepoID: *ctx.RepoID}
	if ctx.StreamID != nil {
		s.StreamID = *ctx.StreamID
	}
	return s
}
