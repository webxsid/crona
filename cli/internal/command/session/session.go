package session

import (
	"errors"
	"fmt"
	"io"

	contextcmd "crona/cli/internal/command/context"
	flagspkg "crona/cli/internal/flags"
	outputpkg "crona/cli/internal/output"
	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

type Deps struct {
	Stdout     io.Writer
	CallKernel func(method string, params, out any) error
}

func TimerUsage() string {
	return "Usage: crona timer <status|start|pause|resume|end> ...\n"
}

func IssueUsage() string {
	return "Usage: crona issue start (--id <issue-id> | --from-context) [--json]\n"
}

func RunTimer(args []string, deps Deps) error {
	if len(args) == 0 || flagspkg.IsHelpArg(args[0]) {
		_, err := fmt.Fprint(deps.Stdout, TimerUsage())
		return err
	}
	contextDeps := contextcmd.Deps{Stdout: deps.Stdout, CallKernel: deps.CallKernel}
	switch args[0] {
	case "status":
		jsonOut := flagspkg.HasJSON(args[1:])
		var out sharedtypes.TimerState
		if err := deps.CallKernel(protocol.MethodTimerGetState, nil, &out); err != nil {
			return err
		}
		if jsonOut {
			return outputpkg.PrintJSON(deps.Stdout, out)
		}
		return outputpkg.PrintTimerStatus(deps.Stdout, out)
	case "start":
		fs := flagspkg.New("timer start")
		issueID := fs.Int64("issue-id", 0, "")
		fromContext := fs.Bool("from-context", false, "")
		jsonOut := fs.Bool("json", false, "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *fromContext {
			resolvedIssueID, err := contextcmd.ResolveIssueID(contextDeps)
			if err != nil {
				return err
			}
			*issueID = resolvedIssueID
		}
		req := shareddto.TimerStartRequest{}
		if *issueID > 0 {
			req.IssueID = issueID
		}
		var out sharedtypes.TimerState
		if err := deps.CallKernel(protocol.MethodTimerStart, req, &out); err != nil {
			return formatStartFocusError(err)
		}
		if *jsonOut {
			return outputpkg.PrintJSON(deps.Stdout, out)
		}
		return outputpkg.PrintTimerResult(deps.Stdout, out, "timer started")
	case "pause":
		jsonOut := flagspkg.HasJSON(args[1:])
		var out sharedtypes.TimerState
		if err := deps.CallKernel(protocol.MethodTimerPause, nil, &out); err != nil {
			return err
		}
		if jsonOut {
			return outputpkg.PrintJSON(deps.Stdout, out)
		}
		return outputpkg.PrintTimerResult(deps.Stdout, out, "timer paused")
	case "resume":
		jsonOut := flagspkg.HasJSON(args[1:])
		var out sharedtypes.TimerState
		if err := deps.CallKernel(protocol.MethodTimerResume, nil, &out); err != nil {
			return err
		}
		if jsonOut {
			return outputpkg.PrintJSON(deps.Stdout, out)
		}
		return outputpkg.PrintTimerResult(deps.Stdout, out, "timer resumed")
	case "end":
		fs := flagspkg.New("timer end")
		jsonOut := fs.Bool("json", false, "")
		commit := fs.String("commit-message", "", "")
		workedOn := fs.String("worked-on", "", "")
		outcome := fs.String("outcome", "", "")
		nextStep := fs.String("next-step", "", "")
		blockers := fs.String("blockers", "", "")
		links := fs.String("links", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		req := shareddto.EndSessionRequest{
			CommitMessage: outputpkg.OptionalFlag(*commit),
			WorkedOn:      outputpkg.OptionalFlag(*workedOn),
			Outcome:       outputpkg.OptionalFlag(*outcome),
			NextStep:      outputpkg.OptionalFlag(*nextStep),
			Blockers:      outputpkg.OptionalFlag(*blockers),
			Links:         outputpkg.OptionalFlag(*links),
		}
		var out sharedtypes.TimerState
		if err := deps.CallKernel(protocol.MethodTimerEnd, req, &out); err != nil {
			return err
		}
		if *jsonOut {
			return outputpkg.PrintJSON(deps.Stdout, out)
		}
		return outputpkg.PrintTimerResult(deps.Stdout, out, "timer ended")
	default:
		return fmt.Errorf("unknown timer command: %s", args[0])
	}
}

func RunIssue(args []string, deps Deps) error {
	if len(args) == 0 || flagspkg.IsHelpArg(args[0]) {
		_, err := fmt.Fprint(deps.Stdout, IssueUsage())
		return err
	}
	switch args[0] {
	case "start":
		fs := flagspkg.New("issue start")
		id := fs.Int64("id", 0, "")
		fromContext := fs.Bool("from-context", false, "")
		jsonOut := fs.Bool("json", false, "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *fromContext {
			resolvedIssueID, err := contextcmd.ResolveIssueID(contextcmd.Deps{Stdout: deps.Stdout, CallKernel: deps.CallKernel})
			if err != nil {
				return err
			}
			*id = resolvedIssueID
		}
		if *id <= 0 {
			return fmt.Errorf("issue id is required")
		}
		var out sharedtypes.TimerState
		if err := deps.CallKernel(protocol.MethodTimerStart, shareddto.TimerStartRequest{IssueID: id}, &out); err != nil {
			return formatStartFocusError(err)
		}
		if *jsonOut {
			return outputpkg.PrintJSON(deps.Stdout, out)
		}
		return outputpkg.PrintTimerResult(deps.Stdout, out, "issue focus started")
	default:
		return fmt.Errorf("unknown issue command: %s", args[0])
	}
}

func formatStartFocusError(err error) error {
	var rpcErr *protocol.RPCError
	if !errors.As(err, &rpcErr) || rpcErr == nil || rpcErr.Code != protocol.ErrorCodeStashConflict {
		return err
	}
	var conflict sharedtypes.StashConflict
	if decodeErr := rpcErr.DecodeData(&conflict); decodeErr != nil || conflict.IssueID == 0 {
		return err
	}
	count := len(conflict.Stashes)
	if count == 1 {
		return fmt.Errorf("cannot start focus: 1 stash exists for issue #%d; resume or drop the stash before starting a new session", conflict.IssueID)
	}
	return fmt.Errorf("cannot start focus: %d stashes exist for issue #%d; resume or drop a stash before starting a new session", count, conflict.IssueID)
}
