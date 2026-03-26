package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"crona/shared/config"
	shareddto "crona/shared/dto"
	"crona/shared/localipc"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

var (
	callKernelFn   = callKernel
	ensureKernelFn = ensureKernel
	runtimeBaseDir = config.RuntimeBaseDir
	readFileFn     = os.ReadFile
	kernelBinaryFn = config.KernelBinaryName
	osExecutableFn = os.Executable
	execLookPathFn = exec.LookPath
	osStatFn       = os.Stat
	osGetwdFn      = os.Getwd
	startKernelFn  = startKernel
	timeSleepFn    = time.Sleep
)

func main() {
	_ = config.Load()
	if err := run(os.Args[1:]); err != nil {
		fail(err.Error())
	}
}

func run(args []string) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		fmt.Print(rootUsage())
		return nil
	}

	switch args[0] {
	case "help":
		fmt.Print(rootUsage())
		return nil
	case "kernel":
		return runKernel(args[1:])
	case "completion":
		return runCompletion(args[1:])
	case "context":
		return runContext(args[1:])
	case "timer":
		return runTimer(args[1:])
	case "issue":
		return runIssue(args[1:])
	case "update":
		return runUpdate(args[1:])
	case "export":
		return runExport(args[1:])
	case "dev":
		return runDev(args[1:])
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func runDev(args []string) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		fmt.Print(devUsage())
		return nil
	}
	jsonOut := hasJSONFlag(args[1:])
	var method string
	switch args[0] {
	case "seed":
		method = protocol.MethodKernelSeedDev
	case "clear":
		method = protocol.MethodKernelClearDev
	default:
		return fmt.Errorf("unknown dev command: %s", args[0])
	}
	if err := callKernelFn(method, nil, nil); err != nil {
		return err
	}
	if jsonOut {
		return printJSON(map[string]any{"ok": true, "command": args[0]})
	}
	fmt.Printf("%s ok\n", args[0])
	return nil
}

func runKernel(args []string) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		fmt.Print(kernelUsage())
		return nil
	}
	jsonOut := hasJSONFlag(args[1:])
	switch args[0] {
	case "attach":
		info, err := ensureKernelFn()
		if err != nil {
			return err
		}
		if jsonOut {
			return printJSON(info)
		}
		fmt.Printf("kernel attached\npid: %d\nendpoint: %s\n", info.PID, kernelEndpoint(info))
		return nil
	case "detach":
		if err := callKernelFn(protocol.MethodKernelShutdown, nil, nil); err != nil {
			return err
		}
		if jsonOut {
			return printJSON(map[string]any{"ok": true})
		}
		fmt.Println("kernel detached")
		return nil
	case "info", "status":
		var out sharedtypes.KernelInfo
		if err := callKernelFn(protocol.MethodKernelInfoGet, nil, &out); err != nil {
			return err
		}
		if jsonOut {
			return printJSON(out)
		}
		fmt.Printf("pid: %d\ntransport: %s\nendpoint: %s\nenv: %s\nstarted: %s\nscratch: %s\n", out.PID, kernelTransport(&out), kernelEndpoint(&out), out.Env, out.StartedAt, out.ScratchDir)
		return nil
	default:
		return fmt.Errorf("unknown kernel command: %s", args[0])
	}
}

func runCompletion(args []string) error {
	if len(args) == 0 || (len(args) == 1 && isHelpArg(args[0])) {
		fmt.Print(completionUsage())
		return nil
	}
	if len(args) != 1 {
		return fmt.Errorf("usage: %s", strings.TrimSpace(completionUsage()))
	}
	switch args[0] {
	case "zsh":
		fmt.Print(zshCompletion())
		return nil
	case "bash":
		fmt.Print(bashCompletion())
		return nil
	case "fish":
		fmt.Print(fishCompletion())
		return nil
	default:
		return fmt.Errorf("unknown shell: %s", args[0])
	}
}

func runContext(args []string) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		fmt.Print(contextUsage())
		return nil
	}
	switch args[0] {
	case "get":
		jsonOut := hasJSONFlag(args[1:])
		var out sharedtypes.ActiveContext
		if err := callKernel(protocol.MethodContextGet, nil, &out); err != nil {
			return err
		}
		if jsonOut {
			return printJSON(out)
		}
		fmt.Printf("repo: %s\nstream: %s\nissue: %s\n", optionalText(out.RepoName, "-"), optionalText(out.StreamName, "-"), optionalText(out.IssueTitle, "-"))
		return nil
	case "clear":
		jsonOut := hasJSONFlag(args[1:])
		if err := callKernel(protocol.MethodContextClear, nil, nil); err != nil {
			return err
		}
		if jsonOut {
			return printJSON(map[string]any{"ok": true})
		}
		fmt.Println("context cleared")
		return nil
	case "set":
		fs := flag.NewFlagSet("context set", flag.ContinueOnError)
		fs.SetOutput(ioDiscard{})
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
		if err := callKernel(protocol.MethodContextSet, req, &out); err != nil {
			return err
		}
		if *jsonOut {
			return printJSON(out)
		}
		fmt.Printf("context set: repo=%s stream=%s issue=%s\n", optionalText(out.RepoName, "-"), optionalText(out.StreamName, "-"), optionalText(out.IssueTitle, "-"))
		return nil
	default:
		return fmt.Errorf("unknown context command: %s", args[0])
	}
}

func runTimer(args []string) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		fmt.Print(timerUsage())
		return nil
	}
	switch args[0] {
	case "status":
		jsonOut := hasJSONFlag(args[1:])
		var out sharedtypes.TimerState
		if err := callKernel(protocol.MethodTimerGetState, nil, &out); err != nil {
			return err
		}
		if jsonOut {
			return printJSON(out)
		}
		segment := "-"
		if out.SegmentType != nil {
			segment = string(*out.SegmentType)
		}
		fmt.Printf("state: %s\nsegment: %s\nelapsed: %ds\n", out.State, segment, out.ElapsedSeconds)
		return nil
	case "start":
		fs := flag.NewFlagSet("timer start", flag.ContinueOnError)
		fs.SetOutput(ioDiscard{})
		issueID := fs.Int64("issue-id", 0, "")
		jsonOut := fs.Bool("json", false, "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		req := struct {
			IssueID *int64 `json:"issueId,omitempty"`
		}{}
		if *issueID > 0 {
			req.IssueID = issueID
		}
		var out sharedtypes.TimerState
		if err := callKernel(protocol.MethodTimerStart, req, &out); err != nil {
			return err
		}
		return printTimerResult(out, *jsonOut, "timer started")
	case "pause":
		jsonOut := hasJSONFlag(args[1:])
		var out sharedtypes.TimerState
		if err := callKernel(protocol.MethodTimerPause, nil, &out); err != nil {
			return err
		}
		return printTimerResult(out, jsonOut, "timer paused")
	case "resume":
		jsonOut := hasJSONFlag(args[1:])
		var out sharedtypes.TimerState
		if err := callKernel(protocol.MethodTimerResume, nil, &out); err != nil {
			return err
		}
		return printTimerResult(out, jsonOut, "timer resumed")
	case "end":
		fs := flag.NewFlagSet("timer end", flag.ContinueOnError)
		fs.SetOutput(ioDiscard{})
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
			CommitMessage: optionalFlag(*commit),
			WorkedOn:      optionalFlag(*workedOn),
			Outcome:       optionalFlag(*outcome),
			NextStep:      optionalFlag(*nextStep),
			Blockers:      optionalFlag(*blockers),
			Links:         optionalFlag(*links),
		}
		var out sharedtypes.TimerState
		if err := callKernel(protocol.MethodTimerEnd, req, &out); err != nil {
			return err
		}
		return printTimerResult(out, *jsonOut, "timer ended")
	default:
		return fmt.Errorf("unknown timer command: %s", args[0])
	}
}

func runIssue(args []string) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		fmt.Print(issueUsage())
		return nil
	}
	switch args[0] {
	case "start":
		fs := flag.NewFlagSet("issue start", flag.ContinueOnError)
		fs.SetOutput(ioDiscard{})
		id := fs.Int64("id", 0, "")
		jsonOut := fs.Bool("json", false, "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *id <= 0 {
			return fmt.Errorf("issue id is required")
		}
		var out sharedtypes.TimerState
		if err := callKernel(protocol.MethodTimerStart, struct {
			IssueID *int64 `json:"issueId,omitempty"`
		}{IssueID: id}, &out); err != nil {
			return err
		}
		return printTimerResult(out, *jsonOut, "issue focus started")
	default:
		return fmt.Errorf("unknown issue command: %s", args[0])
	}
}

func runExport(args []string) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		fmt.Print(exportUsage())
		return nil
	}
	switch args[0] {
	case "calendar":
		fs := flag.NewFlagSet("export calendar", flag.ContinueOnError)
		fs.SetOutput(ioDiscard{})
		repoID := fs.Int64("repo-id", 0, "")
		jsonOut := fs.Bool("json", false, "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		selectedRepoID, err := resolveCalendarRepoID(*repoID)
		if err != nil {
			return err
		}
		req := shareddto.ExportCalendarRequest{RepoID: selectedRepoID}
		var out sharedtypes.CalendarExportResult
		if err := callKernel(protocol.MethodExportCalendar, req, &out); err != nil {
			return err
		}
		if strings.TrimSpace(out.IssuesFilePath) == "" || strings.TrimSpace(out.SessionsFilePath) == "" {
			return errors.New("calendar export response is incomplete; restart the kernel so the updated export handler is loaded")
		}
		if *jsonOut {
			return printJSON(out)
		}
		fmt.Printf("calendar issues export written: %s\n", out.IssuesFilePath)
		fmt.Printf("calendar sessions export written: %s\n", out.SessionsFilePath)
		return nil
	default:
		return fmt.Errorf("unknown export command: %s", args[0])
	}
}

func runUpdate(args []string) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		fmt.Print(updateUsage())
		return nil
	}
	jsonOut := hasJSONFlag(args[1:])
	switch args[0] {
	case "status":
		var out sharedtypes.UpdateStatus
		if err := callKernel(protocol.MethodUpdateStatusGet, nil, &out); err != nil {
			return err
		}
		if jsonOut {
			return printJSON(out)
		}
		return printUpdateStatus(out)
	case "check":
		var out sharedtypes.UpdateStatus
		if err := callKernel(protocol.MethodUpdateCheck, nil, &out); err != nil {
			return err
		}
		if jsonOut {
			return printJSON(out)
		}
		return printUpdateStatus(out)
	case "dismiss":
		var out sharedtypes.UpdateStatus
		if err := callKernel(protocol.MethodUpdateDismiss, nil, &out); err != nil {
			return err
		}
		if jsonOut {
			return printJSON(out)
		}
		if strings.TrimSpace(out.DismissedVersion) == "" {
			fmt.Println("no update dismissed")
			return nil
		}
		fmt.Printf("dismissed update prompt for v%s\n", out.DismissedVersion)
		return nil
	case "notes":
		var out sharedtypes.UpdateStatus
		if err := callKernel(protocol.MethodUpdateStatusGet, nil, &out); err != nil {
			return err
		}
		if jsonOut {
			return printJSON(out)
		}
		return printUpdateNotes(out)
	default:
		return fmt.Errorf("unknown update command: %s", args[0])
	}
}

func resolveCalendarRepoID(explicit int64) (int64, error) {
	if explicit > 0 {
		return explicit, nil
	}
	var ctxOut sharedtypes.ActiveContext
	if err := callKernel(protocol.MethodContextGet, nil, &ctxOut); err == nil && ctxOut.RepoID != nil && *ctxOut.RepoID > 0 {
		return *ctxOut.RepoID, nil
	}
	var repos []sharedtypes.Repo
	if err := callKernel(protocol.MethodRepoList, nil, &repos); err != nil {
		return 0, err
	}
	if len(repos) == 0 {
		return 0, errors.New("calendar export requires at least one repo")
	}
	return repos[0].ID, nil
}

func isHelpArg(value string) bool {
	switch strings.TrimSpace(value) {
	case "-h", "--help":
		return true
	default:
		return false
	}
}

func rootUsage() string {
	return fmt.Sprintf(`Usage: %s <command> [args]

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

func kernelUsage() string {
	return "Usage: crona kernel <attach|detach|info|status> [--json]\n"
}

func completionUsage() string {
	return "Usage: crona completion <zsh|bash|fish>\n"
}

func contextUsage() string {
	return "Usage: crona context <get|set|clear> ...\n"
}

func timerUsage() string {
	return "Usage: crona timer <status|start|pause|resume|end> ...\n"
}

func issueUsage() string {
	return "Usage: crona issue start --id <issue-id> [--json]\n"
}

func updateUsage() string {
	return "Usage: crona update <status|check|dismiss|notes> [--json]\n"
}

func exportUsage() string {
	return "Usage: crona export calendar [--repo-id <id>] [--json]\n"
}

func printUpdateStatus(status sharedtypes.UpdateStatus) error {
	fmt.Printf("current: %s\nenabled: %t\nprompt: %t\n", status.CurrentVersion, status.Enabled, status.PromptEnabled)
	if strings.TrimSpace(status.CheckedAt) != "" {
		fmt.Printf("checked: %s\n", status.CheckedAt)
	}
	if strings.TrimSpace(status.LatestVersion) != "" {
		fmt.Printf("latest: %s\n", status.LatestVersion)
	}
	if strings.TrimSpace(status.ReleaseName) != "" {
		fmt.Printf("title: %s\n", status.ReleaseName)
	} else if summary := firstReleaseSummary(status.ReleaseNotes); summary != "" {
		fmt.Printf("summary: %s\n", summary)
	}
	if strings.TrimSpace(status.ReleaseURL) != "" {
		fmt.Printf("release: %s\n", status.ReleaseURL)
	}
	if status.UpdateAvailable {
		fmt.Println("update: available")
	} else {
		fmt.Println("update: none")
	}
	if strings.TrimSpace(status.DismissedVersion) != "" {
		fmt.Printf("dismissed: %s\n", status.DismissedVersion)
	}
	if strings.TrimSpace(status.Error) != "" {
		fmt.Printf("error: %s\n", status.Error)
	}
	return nil
}

func printUpdateNotes(status sharedtypes.UpdateStatus) error {
	if err := printUpdateStatus(status); err != nil {
		return err
	}
	fmt.Println()
	fmt.Println("release notes:")
	notes := strings.TrimSpace(status.ReleaseNotes)
	if notes == "" {
		fmt.Println("  (no release notes published)")
		return nil
	}
	fmt.Println(notes)
	return nil
}

func firstReleaseSummary(body string) string {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
		if line != "" {
			return line
		}
	}
	return ""
}

func devUsage() string {
	return fmt.Sprintf("Usage: %s dev <seed|clear> [--json]\n", cliCommandName())
}

func callKernel(method string, params, out any) error {
	info, err := readKernelInfo()
	if err != nil {
		return err
	}
	conn, err := localipc.Dial(kernelEndpoint(info), 5*time.Second)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))

	var rawParams json.RawMessage
	if params != nil {
		body, err := json.Marshal(params)
		if err != nil {
			return err
		}
		rawParams = body
	}
	body, err := json.Marshal(protocol.Request{
		ID:     "crona-cli",
		Method: method,
		Params: rawParams,
	})
	if err != nil {
		return err
	}
	if _, err := conn.Write(append(body, '\n')); err != nil {
		return err
	}
	var resp protocol.Response
	if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&resp); err != nil {
		return err
	}
	if resp.Error != nil {
		return fmt.Errorf("%s: %s", resp.Error.Code, resp.Error.Message)
	}
	if out == nil || len(resp.Result) == 0 {
		return nil
	}
	return json.Unmarshal(resp.Result, out)
}

func ensureKernel() (*sharedtypes.KernelInfo, error) {
	if info, err := readKernelInfo(); err == nil {
		if isHealthy(info) {
			return info, nil
		}
	}
	if err := launchKernel(); err != nil {
		return nil, fmt.Errorf("launch kernel: %w", err)
	}
	for i := 0; i < 20; i++ {
		timeSleepFn(250 * time.Millisecond)
		if info, err := readKernelInfo(); err == nil && isHealthy(info) {
			return info, nil
		}
	}
	return nil, fmt.Errorf("kernel failed to start within 5s")
}

func readKernelInfo() (*sharedtypes.KernelInfo, error) {
	base, err := runtimeBaseDir()
	if err != nil {
		return nil, err
	}
	body, err := readFileFn(filepath.Join(base, "kernel.json"))
	if err != nil {
		return nil, err
	}
	var info sharedtypes.KernelInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}
	normalizeKernelInfo(&info)
	return &info, nil
}

func isHealthy(info *sharedtypes.KernelInfo) bool {
	if info == nil || strings.TrimSpace(kernelEndpoint(info)) == "" {
		return false
	}
	conn, err := localipc.Dial(kernelEndpoint(info), 2*time.Second)
	if err != nil {
		return false
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	body, err := json.Marshal(protocol.Request{ID: "healthcheck", Method: protocol.MethodHealthGet})
	if err != nil {
		return false
	}
	if _, err := conn.Write(append(body, '\n')); err != nil {
		return false
	}
	var resp protocol.Response
	if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&resp); err != nil {
		return false
	}
	return resp.Error == nil
}

func normalizeKernelInfo(info *sharedtypes.KernelInfo) {
	if info == nil {
		return
	}
	if strings.TrimSpace(info.Transport) == "" {
		info.Transport = localipc.DefaultTransport()
	}
	if strings.TrimSpace(info.Endpoint) == "" {
		info.Endpoint = strings.TrimSpace(info.SocketPath)
	}
	if strings.TrimSpace(info.SocketPath) == "" && info.Transport == localipc.TransportUnixSocket {
		info.SocketPath = info.Endpoint
	}
}

func kernelEndpoint(info *sharedtypes.KernelInfo) string {
	if info == nil {
		return ""
	}
	if strings.TrimSpace(info.Endpoint) != "" {
		return info.Endpoint
	}
	return strings.TrimSpace(info.SocketPath)
}

func kernelTransport(info *sharedtypes.KernelInfo) string {
	if info == nil || strings.TrimSpace(info.Transport) == "" {
		return localipc.DefaultTransport()
	}
	return info.Transport
}

type launchCandidate struct {
	name string
	cmd  string
	args []string
	dir  string
}

func launchKernel() error {
	candidates := kernelLaunchCandidates()
	if len(candidates) == 0 {
		return errors.New("no kernel launcher candidates found")
	}
	failures := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if err := startKernelFn(candidate); err == nil {
			return nil
		} else {
			failures = append(failures, fmt.Sprintf("%s: %v", candidate.name, err))
		}
	}
	return errors.New(strings.Join(failures, "; "))
}

func kernelLaunchCandidates() []launchCandidate {
	candidates := make([]launchCandidate, 0, 3)
	seen := make(map[string]struct{})
	add := func(candidate launchCandidate) {
		key := candidate.cmd + "\x00" + strings.Join(candidate.args, "\x00") + "\x00" + candidate.dir
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		candidates = append(candidates, candidate)
	}
	if exe, err := osExecutableFn(); err == nil {
		kernelName := kernelBinaryFn()
		sibling := filepath.Join(filepath.Dir(exe), kernelName)
		if info, err := osStatFn(sibling); err == nil && !info.IsDir() {
			add(launchCandidate{name: "sibling " + kernelName, cmd: sibling})
		}
	}
	if pathCmd, err := execLookPathFn(kernelBinaryFn()); err == nil {
		add(launchCandidate{name: "PATH " + kernelBinaryFn(), cmd: pathCmd})
	}
	if repoRoot, err := findRepoRoot(); err == nil {
		repoBin := filepath.Join(repoRoot, "bin", kernelBinaryFn())
		if info, err := osStatFn(repoBin); err == nil && !info.IsDir() {
			add(launchCandidate{name: "repo bin " + kernelBinaryFn(), cmd: repoBin})
		}
		if _, err := osStatFn(filepath.Join(repoRoot, "kernel", "cmd", "crona-kernel")); err == nil {
			if goCmd, lookErr := execLookPathFn("go"); lookErr == nil {
				add(launchCandidate{name: "repo-local go run", cmd: goCmd, args: []string{"run", "./kernel/cmd/crona-kernel"}, dir: repoRoot})
			}
		}
	}
	return candidates
}

func findRepoRoot() (string, error) {
	starts := make([]string, 0, 2)
	if wd, err := osGetwdFn(); err == nil {
		starts = append(starts, wd)
	}
	if exe, err := osExecutableFn(); err == nil {
		starts = append(starts, filepath.Dir(exe))
	}
	seen := make(map[string]struct{})
	for _, start := range starts {
		dir := start
		for {
			if _, ok := seen[dir]; ok {
				break
			}
			seen[dir] = struct{}{}
			if fileExists(filepath.Join(dir, "go.work")) && fileExists(filepath.Join(dir, "kernel", "cmd", "crona-kernel")) {
				return dir, nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return "", errors.New("repo root not found")
}

func fileExists(path string) bool {
	info, err := osStatFn(path)
	return err == nil && !info.IsDir()
}

func startKernel(candidate launchCandidate) error {
	cmd := exec.Command(candidate.cmd, candidate.args...)
	cmd.Dir = candidate.dir
	cmd.Stdin = nil
	cmd.Stdout = nil
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	waitCh := make(chan error, 1)
	go func() { waitCh <- cmd.Wait() }()
	select {
	case err := <-waitCh:
		detail := strings.TrimSpace(stderr.String())
		if detail == "" && err != nil {
			detail = err.Error()
		}
		if detail == "" {
			detail = "exited immediately"
		}
		return errors.New(detail)
	case <-time.After(300 * time.Millisecond):
		return nil
	}
}

func zshCompletion() string {
	name := cliCommandName()
	return fmt.Sprintf(`#compdef %s
_%s() {
  local -a commands
  commands=('kernel:Kernel commands' 'completion:Shell completions' 'context:Context commands' 'timer:Timer commands' 'issue:Issue commands' 'update:Update commands' 'export:Export commands' 'dev:Dev-only commands')
  if (( CURRENT == 2 )); then
    _describe 'command' commands
    return
  fi
  case "${words[2]}" in
    kernel) _values 'kernel command' attach detach info status ;;
    completion) _values 'shell' zsh bash fish ;;
    context) _values 'context command' get set clear ;;
    timer) _values 'timer command' status start pause resume end ;;
    issue) _values 'issue command' start ;;
    update) _values 'update command' status check dismiss notes ;;
    export) _values 'export command' calendar ;;
    dev) _values 'dev command' seed clear ;;
  esac
}
_%s "$@"
`, name, name, name)
}

func bashCompletion() string {
	name := cliCommandName()
	return fmt.Sprintf(`_%s()
{
  local cur prev words cword
  _init_completion || return
  if [[ ${cword} -eq 1 ]]; then
    COMPREPLY=( $(compgen -W "kernel completion context timer issue update export dev" -- "$cur") )
    return
  fi
  case "${words[1]}" in
    kernel) COMPREPLY=( $(compgen -W "attach detach info status" -- "$cur") ) ;;
    completion) COMPREPLY=( $(compgen -W "zsh bash fish" -- "$cur") ) ;;
    context) COMPREPLY=( $(compgen -W "get set clear" -- "$cur") ) ;;
    timer) COMPREPLY=( $(compgen -W "status start pause resume end" -- "$cur") ) ;;
    issue) COMPREPLY=( $(compgen -W "start" -- "$cur") ) ;;
    update) COMPREPLY=( $(compgen -W "status check dismiss notes" -- "$cur") ) ;;
    export) COMPREPLY=( $(compgen -W "calendar" -- "$cur") ) ;;
    dev) COMPREPLY=( $(compgen -W "seed clear" -- "$cur") ) ;;
  esac
}
complete -F _%s %s
`, name, name, name)
}

func fishCompletion() string {
	name := cliCommandName()
	return fmt.Sprintf(`complete -c %s -f -n "__fish_use_subcommand" -a "kernel completion context timer issue update export dev"
complete -c %s -f -n "__fish_seen_subcommand_from kernel" -a "attach detach info status"
complete -c %s -f -n "__fish_seen_subcommand_from completion" -a "zsh bash fish"
complete -c %s -f -n "__fish_seen_subcommand_from context" -a "get set clear"
complete -c %s -f -n "__fish_seen_subcommand_from timer" -a "status start pause resume end"
complete -c %s -f -n "__fish_seen_subcommand_from issue" -a "start"
complete -c %s -f -n "__fish_seen_subcommand_from update" -a "status check dismiss notes"
complete -c %s -f -n "__fish_seen_subcommand_from export" -a "calendar"
complete -c %s -f -n "__fish_seen_subcommand_from dev" -a "seed clear"
`, name, name, name, name, name, name, name, name, name)
}

func cliCommandName() string {
	name := filepath.Base(os.Args[0])
	if strings.TrimSpace(name) != "" && name != "." {
		return name
	}
	return config.CLIBinaryName()
}

func printTimerResult(out sharedtypes.TimerState, jsonOut bool, message string) error {
	if jsonOut {
		return printJSON(out)
	}
	fmt.Printf("%s\nstate: %s\nelapsed: %ds\n", message, out.State, out.ElapsedSeconds)
	return nil
}

func printJSON(value any) error {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(body))
	return nil
}

func optionalFlag(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func optionalText(value *string, fallback string) string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return fallback
	}
	return strings.TrimSpace(*value)
}

func hasJSONFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--json" {
			return true
		}
	}
	return false
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) {
	return len(p), nil
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
