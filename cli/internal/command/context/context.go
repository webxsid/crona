package context

import (
	"errors"
	"fmt"
	"io"

	"crona/cli/internal/flags"
	"crona/cli/internal/output"
	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

type Deps struct {
	Stdout     io.Writer
	CallKernel func(method string, params, out any) error
}

func Usage() string {
	return "Usage: crona context <get|set|clear|clear-issue|switch-repo|switch-stream|switch-issue> ...\n"
}

func Run(args []string, deps Deps) error {
	if len(args) == 0 || flags.IsHelpArg(args[0]) {
		_, err := fmt.Fprint(deps.Stdout, Usage())
		return err
	}
	switch args[0] {
	case "get":
		jsonOut := flags.HasJSON(args[1:])
		var out sharedtypes.ActiveContext
		if err := deps.CallKernel(protocol.MethodContextGet, nil, &out); err != nil {
			return err
		}
		if jsonOut {
			return output.PrintJSON(deps.Stdout, out)
		}
		return output.PrintContext(deps.Stdout, out)
	case "clear":
		jsonOut := flags.HasJSON(args[1:])
		if err := deps.CallKernel(protocol.MethodContextClear, nil, nil); err != nil {
			return err
		}
		if jsonOut {
			return output.PrintJSON(deps.Stdout, map[string]any{"ok": true})
		}
		_, err := fmt.Fprintln(deps.Stdout, "context cleared")
		return err
	case "set":
		fs := flags.New("context set")
		repoID := fs.Int64("repo-id", 0, "")
		streamID := fs.Int64("stream-id", 0, "")
		issueID := fs.Int64("issue-id", 0, "")
		jsonOut := fs.Bool("json", false, "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		req := shareddto.UpdateContextRequest{}
		if *repoID > 0 {
			req.RepoID = repoID
		}
		if *streamID > 0 {
			req.StreamID = streamID
		}
		if *issueID > 0 {
			req.IssueID = issueID
		}
		var out sharedtypes.ActiveContext
		if err := deps.CallKernel(protocol.MethodContextSet, req, &out); err != nil {
			return err
		}
		if *jsonOut {
			return output.PrintJSON(deps.Stdout, out)
		}
		return output.PrintContextInline(deps.Stdout, "context set", out)
	case "clear-issue":
		jsonOut := flags.HasJSON(args[1:])
		if err := deps.CallKernel(protocol.MethodContextClearIssue, nil, nil); err != nil {
			return err
		}
		if jsonOut {
			return output.PrintJSON(deps.Stdout, map[string]any{"ok": true})
		}
		_, err := fmt.Fprintln(deps.Stdout, "context issue cleared")
		return err
	case "switch-repo":
		return runSwitchRepo(args[1:], deps)
	case "switch-stream":
		return runSwitchStream(args[1:], deps)
	case "switch-issue":
		return runSwitchIssue(args[1:], deps)
	default:
		return fmt.Errorf("unknown context command: %s", args[0])
	}
}

func ResolveActive(deps Deps) (*sharedtypes.ActiveContext, error) {
	var out sharedtypes.ActiveContext
	if err := deps.CallKernel(protocol.MethodContextGet, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func ResolveIssueID(deps Deps) (int64, error) {
	ctx, err := ResolveActive(deps)
	if err != nil {
		return 0, err
	}
	if ctx.IssueID == nil || *ctx.IssueID <= 0 {
		return 0, errors.New("current context has no checked-out issue")
	}
	return *ctx.IssueID, nil
}

func runSwitchRepo(args []string, deps Deps) error {
	fs := flags.New("context switch-repo")
	id := fs.Int64("id", 0, "")
	jsonOut := fs.Bool("json", false, "")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id <= 0 {
		return errors.New("repo id is required")
	}
	var out sharedtypes.ActiveContext
	if err := deps.CallKernel(protocol.MethodContextSwitchRepo, shareddto.SwitchRepoRequest{RepoID: *id}, &out); err != nil {
		return err
	}
	if *jsonOut {
		return output.PrintJSON(deps.Stdout, out)
	}
	return output.PrintContextInline(deps.Stdout, "context switched", out)
}

func runSwitchStream(args []string, deps Deps) error {
	fs := flags.New("context switch-stream")
	id := fs.Int64("id", 0, "")
	jsonOut := fs.Bool("json", false, "")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id <= 0 {
		return errors.New("stream id is required")
	}
	var out sharedtypes.ActiveContext
	if err := deps.CallKernel(protocol.MethodContextSwitchStream, shareddto.SwitchStreamRequest{StreamID: *id}, &out); err != nil {
		return err
	}
	if *jsonOut {
		return output.PrintJSON(deps.Stdout, out)
	}
	return output.PrintContextInline(deps.Stdout, "context switched", out)
}

func runSwitchIssue(args []string, deps Deps) error {
	fs := flags.New("context switch-issue")
	id := fs.Int64("id", 0, "")
	jsonOut := fs.Bool("json", false, "")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id <= 0 {
		return errors.New("issue id is required")
	}
	var out sharedtypes.ActiveContext
	if err := deps.CallKernel(protocol.MethodContextSwitchIssue, shareddto.SwitchIssueRequest{IssueID: *id}, &out); err != nil {
		return err
	}
	if *jsonOut {
		return output.PrintJSON(deps.Stdout, out)
	}
	return output.PrintContextInline(deps.Stdout, "context switched", out)
}
