package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	completioncmd "crona/cli/internal/command/completion"
	contextcmd "crona/cli/internal/command/context"
	devcmd "crona/cli/internal/command/dev"
	exportcmd "crona/cli/internal/command/export"
	kernelcmd "crona/cli/internal/command/kernel"
	sessioncmd "crona/cli/internal/command/session"
	updatecmd "crona/cli/internal/command/update"
	flagspkg "crona/cli/internal/flags"
	runtimepkg "crona/cli/internal/runtime"
	"crona/shared/config"
	sharedtypes "crona/shared/types"
)

var (
	runtimeBaseDir = config.RuntimeBaseDir
	readFileFn     = os.ReadFile
	kernelBinaryFn = config.KernelBinaryName
	tuiBinaryFn    = config.TUIBinaryName
	osExecutableFn = os.Executable
	execLookPathFn = exec.LookPath
	osStatFn       = os.Stat
	osGetwdFn      = os.Getwd
	startKernelFn  = runtimepkg.StartKernelProcess
	timeSleepFn    = time.Sleep

	callKernelFn = func(method string, params, out any) error {
		return runtimepkg.CallKernel(runtimeDeps(), method, params, out)
	}
	ensureKernelFn = func() (*sharedtypes.KernelInfo, error) {
		return runtimepkg.EnsureKernel(runtimeDeps())
	}
	launchTUIFn = func() error {
		return runtimepkg.LaunchTUI(runtimeDeps())
	}
)

func main() {
	_ = config.Load()
	if err := run(os.Args[1:]); err != nil {
		fail(err.Error())
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return launchTUIFn()
	}
	if flagspkg.IsHelpArg(args[0]) {
		fmt.Print(rootUsage())
		return nil
	}

	switch args[0] {
	case "help":
		fmt.Print(rootUsage())
		return nil
	case "kernel":
		return kernelcmd.Run(args[1:], kernelcmd.Deps{
			Stdout:       os.Stdout,
			CallKernel:   callKernelFn,
			EnsureKernel: ensureKernelFn,
		})
	case "completion":
		return completioncmd.Run(args[1:], cliCommandName(), os.Stdout)
	case "context":
		return contextcmd.Run(args[1:], contextcmd.Deps{
			Stdout:     os.Stdout,
			CallKernel: callKernelFn,
		})
	case "timer":
		return sessioncmd.RunTimer(args[1:], sessioncmd.Deps{
			Stdout:     os.Stdout,
			CallKernel: callKernelFn,
		})
	case "issue":
		return sessioncmd.RunIssue(args[1:], sessioncmd.Deps{
			Stdout:     os.Stdout,
			CallKernel: callKernelFn,
		})
	case "update":
		return updatecmd.Run(args[1:], updatecmd.Deps{
			Stdout:     os.Stdout,
			CallKernel: callKernelFn,
		})
	case "export":
		return exportcmd.Run(args[1:], exportcmd.Deps{
			Stdout:     os.Stdout,
			CallKernel: callKernelFn,
		})
	case "dev":
		return devcmd.Run(args[1:], cliCommandName(), devcmd.Deps{
			Stdout:     os.Stdout,
			CallKernel: callKernelFn,
		})
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func runtimeDeps() runtimepkg.Deps {
	return runtimepkg.Deps{
		RuntimeBaseDir: runtimeBaseDir,
		ReadFile:       readFileFn,
		KernelBinary:   kernelBinaryFn,
		TUIBinary:      tuiBinaryFn,
		OSExecutable:   osExecutableFn,
		ExecLookPath:   execLookPathFn,
		OSStat:         osStatFn,
		OSGetwd:        osGetwdFn,
		StartKernel:    startKernelFn,
		TimeSleep:      timeSleepFn,
	}
}

func rootUsage() string {
	return fmt.Sprintf(`Usage: %s [command] [args]

Run without a command to open the TUI.

Commands:
  help
  kernel      Attach, detach, and inspect the local kernel
  completion  Generate shell completions
  context     Inspect or update checked-out context
  timer       Control the active timer/session
  issue       Start issue focus
  update      Inspect release updates and release notes
  export      Export automation artifacts such as calendar ICS files
  dev         Seed or clear dev data
`, cliCommandName())
}

func cliCommandName() string {
	name := filepath.Base(os.Args[0])
	if strings.TrimSpace(name) != "" && name != "." {
		return name
	}
	return config.CLIBinaryName()
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
