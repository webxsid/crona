// Package capture implements `crona issue add`: parse `- ` lines into issues, show the
// parse, and import it after confirmation.
package capture

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"crona/cli/internal/flags"
	"crona/cli/internal/output"
	shareddto "crona/shared/dto"
	"crona/shared/protocol"
)

type Deps struct {
	Stdout     io.Writer
	Stdin      io.Reader
	CallKernel func(method string, params, out any) error
	// Confirm asks a yes/no question on the terminal. It fails when there is none.
	Confirm func(prompt string) (bool, error)
}

func Usage() string {
	return "Usage: crona issue add (\"text\" | -) [--repo <name>] [--stream <name>] [--yes] [--dry-run]\n" +
		"  \"text\"           One issue line, with optional #repo/stream ~45m @friday +tag annotations\n" +
		"  -                Read `- ` lines from stdin; indented lines become descriptions\n" +
		"  --repo, --stream Default target for lines without #; otherwise the Inbox stream\n" +
		"  --yes            Import every unflagged row without asking\n" +
		"  --dry-run        Print the parse as JSON and import nothing\n"
}

func Run(args []string, deps Deps) error {
	if len(args) == 0 || flags.IsHelpArg(args[0]) {
		_, err := fmt.Fprint(deps.Stdout, Usage())
		return err
	}
	fs := flags.New("issue add")
	repo := fs.String("repo", "", "")
	stream := fs.String("stream", "", "")
	yes := fs.Bool("yes", false, "")
	dryRun := fs.Bool("dry-run", false, "")
	positionals, err := flags.ParseWithPositionals(fs, args)
	if err != nil {
		return err
	}
	text, err := readText(positionals, deps.Stdin)
	if err != nil {
		return err
	}
	req := shareddto.IssueCaptureParseRequest{Text: text, Target: target(*repo, *stream)}
	var proposals []shareddto.IssueCaptureProposal
	if err := deps.CallKernel(protocol.MethodIssueCaptureParse, req, &proposals); err != nil {
		return err
	}
	if *dryRun {
		return output.PrintJSON(deps.Stdout, proposals)
	}
	if len(proposals) == 0 {
		return errors.New("no `- ` lines to import")
	}
	ready := make([]shareddto.IssueCaptureProposal, 0, len(proposals))
	for _, p := range proposals {
		if len(p.Flags) == 0 {
			ready = append(ready, p)
		}
	}
	if err := printRows(deps.Stdout, proposals); err != nil {
		return err
	}
	if len(ready) == 0 {
		return errors.New("every row is flagged; fix the lines and try again")
	}
	if !*yes {
		ok, err := deps.Confirm(fmt.Sprintf("Import %s? [Y/n] ", plural(len(ready), "issue")))
		if err != nil {
			return fmt.Errorf("%w; pass --yes to import without asking", err)
		}
		if !ok {
			_, err := fmt.Fprintln(deps.Stdout, "Nothing imported.")
			return err
		}
	}
	var result shareddto.IssueCaptureImportResult
	if err := deps.CallKernel(protocol.MethodIssueCaptureImport, shareddto.IssueCaptureImportRequest{Proposals: ready}, &result); err != nil {
		return err
	}
	summary := "Imported " + plural(len(result.Issues), "issue")
	if left := len(proposals) - len(ready); left > 0 {
		summary += fmt.Sprintf(" · %s left out", plural(left, "flagged row"))
	}
	_, err = fmt.Fprintln(deps.Stdout, summary)
	return err
}

// readText takes one line from the arguments, or every line from stdin for `-`.
func readText(positionals []string, stdin io.Reader) (string, error) {
	if len(positionals) == 1 && positionals[0] == "-" {
		raw, err := io.ReadAll(stdin)
		return string(raw), err
	}
	line := strings.TrimSpace(strings.Join(positionals, " "))
	if line == "" {
		return "", errors.New("issue text is required")
	}
	return "- " + strings.TrimPrefix(line, "- "), nil
}

func target(repo, stream string) string {
	repo, stream = strings.TrimSpace(repo), strings.TrimSpace(stream)
	if repo != "" && stream != "" {
		return repo + "/" + stream
	}
	return repo + stream
}

func printRows(w io.Writer, proposals []shareddto.IssueCaptureProposal) error {
	for i, p := range proposals {
		marker := fmt.Sprintf("%3d", i+1)
		place := p.RepoName + "/" + p.StreamName
		if len(p.Flags) > 0 {
			marker = "  ?"
			place = strings.Join(p.Flags, ", ")
		}
		fields := []string{place}
		if p.EstimateMinutes != nil {
			fields = append(fields, formatMinutes(*p.EstimateMinutes))
		}
		if p.TodoForDate != nil {
			fields = append(fields, *p.TodoForDate)
		}
		for _, tag := range p.Tags {
			fields = append(fields, "+"+tag)
		}
		if _, err := fmt.Fprintf(w, "%s  %-32s  %s\n", marker, p.Title, strings.Join(fields, "  ")); err != nil {
			return err
		}
	}
	return nil
}

func formatMinutes(minutes int) string {
	switch {
	case minutes >= 60 && minutes%60 == 0:
		return fmt.Sprintf("%dh", minutes/60)
	case minutes >= 60:
		return fmt.Sprintf("%dh%dm", minutes/60, minutes%60)
	}
	return fmt.Sprintf("%dm", minutes)
}

func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

// ConfirmOnTerminal asks on the controlling terminal, so it still works when stdin is
// the piped issue text. Enter alone means yes.
func ConfirmOnTerminal(prompt string) (bool, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return false, errors.New("no terminal to confirm on")
	}
	defer func() { _ = tty.Close() }()
	if _, err := fmt.Fprint(tty, prompt); err != nil {
		return false, err
	}
	answer, err := bufio.NewReader(tty).ReadString('\n')
	if err != nil && answer == "" {
		return false, err
	}
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "", "y", "yes":
		return true, nil
	}
	return false, nil
}
