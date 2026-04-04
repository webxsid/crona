package kernel

import (
	"errors"
	"fmt"
	"io"

	"crona/cli/internal/flags"
	"crona/cli/internal/output"
	runtimepkg "crona/cli/internal/runtime"
	shareddto "crona/shared/dto"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

type Deps struct {
	Stdout       io.Writer
	CallKernel   func(method string, params, out any) error
	EnsureKernel func() (*sharedtypes.KernelInfo, error)
}

func Usage() string {
	return "Usage: crona kernel <attach|detach|restart|wipe-data|info|status> [--json]\n"
}

func Run(args []string, deps Deps) error {
	if len(args) == 0 || flags.IsHelpArg(args[0]) {
		_, err := fmt.Fprint(deps.Stdout, Usage())
		return err
	}
	jsonOut := flags.HasJSON(args[1:])
	switch args[0] {
	case "attach":
		info, err := deps.EnsureKernel()
		if err != nil {
			return err
		}
		if jsonOut {
			return output.PrintJSON(deps.Stdout, info)
		}
		return output.PrintKernelAttach(deps.Stdout, info, runtimepkg.KernelEndpoint(info))
	case "detach":
		if err := deps.CallKernel(protocol.MethodKernelShutdown, nil, nil); err != nil {
			return err
		}
		if jsonOut {
			return output.PrintJSON(deps.Stdout, map[string]any{"ok": true})
		}
		_, err := fmt.Fprintln(deps.Stdout, "kernel detached")
		return err
	case "restart":
		if err := deps.CallKernel(protocol.MethodKernelRestart, nil, nil); err != nil {
			return err
		}
		if jsonOut {
			return output.PrintJSON(deps.Stdout, map[string]any{"ok": true})
		}
		_, err := fmt.Fprintln(deps.Stdout, "kernel restart requested")
		return err
	case "wipe-data":
		fs := flags.New("kernel wipe-data")
		force := fs.Bool("force", false, "")
		jsonFlag := fs.Bool("json", false, "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if !*force {
			return errors.New("kernel wipe-data requires --force")
		}
		if err := deps.CallKernel(protocol.MethodKernelWipeData, shareddto.ConfirmDangerousActionRequest{Confirm: true}, nil); err != nil {
			return err
		}
		if *jsonFlag {
			return output.PrintJSON(deps.Stdout, map[string]any{"ok": true})
		}
		_, err := fmt.Fprintln(deps.Stdout, "runtime data wiped")
		return err
	case "info", "status":
		var out sharedtypes.KernelInfo
		if err := deps.CallKernel(protocol.MethodKernelInfoGet, nil, &out); err != nil {
			return err
		}
		if jsonOut {
			return output.PrintJSON(deps.Stdout, out)
		}
		return output.PrintKernelInfo(deps.Stdout, out, runtimepkg.KernelTransport(&out), runtimepkg.KernelEndpoint(&out))
	default:
		return fmt.Errorf("unknown kernel command: %s", args[0])
	}
}
