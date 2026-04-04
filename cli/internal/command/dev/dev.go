package dev

import (
	"fmt"
	"io"

	flagspkg "crona/cli/internal/flags"
	outputpkg "crona/cli/internal/output"
	"crona/shared/protocol"
)

type Deps struct {
	Stdout     io.Writer
	CallKernel func(method string, params, out any) error
}

func Usage(cliName string) string {
	return fmt.Sprintf("Usage: %s dev <seed|clear> [--json]\n", cliName)
}

func Run(args []string, cliName string, deps Deps) error {
	if len(args) == 0 || flagspkg.IsHelpArg(args[0]) {
		_, err := fmt.Fprint(deps.Stdout, Usage(cliName))
		return err
	}
	jsonOut := flagspkg.HasJSON(args[1:])
	var method string
	switch args[0] {
	case "seed":
		method = protocol.MethodKernelSeedDev
	case "clear":
		method = protocol.MethodKernelClearDev
	default:
		return fmt.Errorf("unknown dev command: %s", args[0])
	}
	if err := deps.CallKernel(method, nil, nil); err != nil {
		return err
	}
	if jsonOut {
		return outputpkg.PrintJSON(deps.Stdout, map[string]any{"ok": true, "command": args[0]})
	}
	_, err := fmt.Fprintf(deps.Stdout, "%s ok\n", args[0])
	return err
}
