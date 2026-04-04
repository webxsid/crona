package update

import (
	"fmt"
	"io"
	"strings"

	flagspkg "crona/cli/internal/flags"
	outputpkg "crona/cli/internal/output"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

type Deps struct {
	Stdout     io.Writer
	CallKernel func(method string, params, out any) error
}

func Usage() string {
	return "Usage: crona update <status|check|dismiss|notes> [--json]\n"
}

func Run(args []string, deps Deps) error {
	if len(args) == 0 || flagspkg.IsHelpArg(args[0]) {
		_, err := fmt.Fprint(deps.Stdout, Usage())
		return err
	}
	jsonOut := flagspkg.HasJSON(args[1:])
	switch args[0] {
	case "status":
		var out sharedtypes.UpdateStatus
		if err := deps.CallKernel(protocol.MethodUpdateStatusGet, nil, &out); err != nil {
			return err
		}
		if jsonOut {
			return outputpkg.PrintJSON(deps.Stdout, out)
		}
		return outputpkg.PrintUpdateStatus(deps.Stdout, out)
	case "check":
		var out sharedtypes.UpdateStatus
		if err := deps.CallKernel(protocol.MethodUpdateCheck, nil, &out); err != nil {
			return err
		}
		if jsonOut {
			return outputpkg.PrintJSON(deps.Stdout, out)
		}
		return outputpkg.PrintUpdateStatus(deps.Stdout, out)
	case "dismiss":
		var out sharedtypes.UpdateStatus
		if err := deps.CallKernel(protocol.MethodUpdateDismiss, nil, &out); err != nil {
			return err
		}
		if jsonOut {
			return outputpkg.PrintJSON(deps.Stdout, out)
		}
		if strings.TrimSpace(out.DismissedVersion) == "" {
			_, err := fmt.Fprintln(deps.Stdout, "no update dismissed")
			return err
		}
		_, err := fmt.Fprintf(deps.Stdout, "dismissed update prompt for v%s\n", out.DismissedVersion)
		return err
	case "notes":
		var out sharedtypes.UpdateStatus
		if err := deps.CallKernel(protocol.MethodUpdateStatusGet, nil, &out); err != nil {
			return err
		}
		if jsonOut {
			return outputpkg.PrintJSON(deps.Stdout, out)
		}
		return outputpkg.PrintUpdateNotes(deps.Stdout, out)
	default:
		return fmt.Errorf("unknown update command: %s", args[0])
	}
}
