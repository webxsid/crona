package tag

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

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
	return "Usage: crona tag <ls|rename|color|delete|prune> ...\n" +
		"  ls [--json]                       List tags with usage counts\n" +
		"  rename <old> <new> [--merge]      Rename a tag; --merge folds it into an existing one\n" +
		"  color <name> <color|none>         Set or clear a tag's color\n" +
		"  delete <name>                     Delete a tag and remove it from every item\n" +
		"  prune                             Delete tags no issue or habit uses\n"
}

// ItemUsage is the usage line for `crona issue tag` and `crona habit tag`.
func ItemUsage(kind string) string {
	return fmt.Sprintf("Usage: crona %s tag <%s-id> [+add ...] [-remove ...] [--json]\n", kind, kind)
}

func Run(args []string, deps Deps) error {
	if len(args) == 0 || flags.IsHelpArg(args[0]) {
		_, err := fmt.Fprint(deps.Stdout, Usage())
		return err
	}
	switch args[0] {
	case "ls", "list":
		var out []sharedtypes.Tag
		if err := deps.CallKernel(protocol.MethodTagList, nil, &out); err != nil {
			return err
		}
		if flags.HasJSON(args[1:]) {
			return output.PrintJSON(deps.Stdout, out)
		}
		return printTags(deps.Stdout, out)
	case "rename":
		fs := flags.New("tag rename")
		merge := fs.Bool("merge", false, "")
		names, err := flags.ParseWithPositionals(fs, args[1:])
		if err != nil {
			return err
		}
		if len(names) != 2 {
			return errors.New("usage: crona tag rename <old> <new> [--merge]")
		}
		req := shareddto.TagRenameRequest{From: names[0], To: names[1], Merge: *merge}
		if err := deps.CallKernel(protocol.MethodTagRename, req, nil); err != nil {
			return err
		}
		_, err = fmt.Fprintf(deps.Stdout, "renamed +%s to +%s\n", strings.TrimPrefix(names[0], "+"), strings.TrimPrefix(names[1], "+"))
		return err
	case "color":
		if len(args) != 3 {
			return fmt.Errorf("usage: crona tag color <name> <%s|none>", strings.Join(sharedtypes.TagColors, "|"))
		}
		req := shareddto.TagSetColorRequest{Name: args[1]}
		if args[2] != "none" {
			color := args[2]
			req.Color = &color
		}
		if err := deps.CallKernel(protocol.MethodTagSetColor, req, nil); err != nil {
			return err
		}
		_, err := fmt.Fprintf(deps.Stdout, "+%s color set to %s\n", strings.TrimPrefix(args[1], "+"), args[2])
		return err
	case "delete", "rm":
		if len(args) != 2 {
			return errors.New("usage: crona tag delete <name>")
		}
		if err := deps.CallKernel(protocol.MethodTagDelete, shareddto.TagNameRequest{Name: args[1]}, nil); err != nil {
			return err
		}
		_, err := fmt.Fprintf(deps.Stdout, "deleted +%s\n", strings.TrimPrefix(args[1], "+"))
		return err
	case "prune":
		var out shareddto.TagRemoveUnusedResponse
		if err := deps.CallKernel(protocol.MethodTagRemoveUnused, nil, &out); err != nil {
			return err
		}
		if flags.HasJSON(args[1:]) {
			return output.PrintJSON(deps.Stdout, out)
		}
		if len(out.Removed) == 0 {
			_, err := fmt.Fprintln(deps.Stdout, "no unused tags")
			return err
		}
		_, err := fmt.Fprintf(deps.Stdout, "removed %s\n", joinTags(out.Removed))
		return err
	default:
		return fmt.Errorf("unknown tag command: %s", args[0])
	}
}

// RunItem adds and removes tags on one issue or habit: `<id> +urgent -waiting`.
func RunItem(kind string, args []string, deps Deps) error {
	if len(args) == 0 || flags.IsHelpArg(args[0]) {
		_, err := fmt.Fprint(deps.Stdout, ItemUsage(kind))
		return err
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil || id <= 0 {
		return fmt.Errorf("%s id is required", kind)
	}
	req := shareddto.UpdateTagsRequest{IDs: []int64{id}}
	jsonOut := false
	for _, arg := range args[1:] {
		switch {
		case arg == "--json":
			jsonOut = true
		case strings.HasPrefix(arg, "-") && len(arg) > 1:
			req.Remove = append(req.Remove, arg[1:])
		case strings.HasPrefix(arg, "+") && len(arg) > 1:
			req.Add = append(req.Add, arg[1:])
		default:
			return fmt.Errorf("tag %q must start with + (add) or - (remove)", arg)
		}
	}
	if len(req.Add) == 0 && len(req.Remove) == 0 {
		return errors.New("give at least one +tag or -tag")
	}
	method := protocol.MethodIssueTagsUpdate
	if kind == "habit" {
		method = protocol.MethodHabitTagsUpdate
	}
	var out json.RawMessage
	if err := deps.CallKernel(method, req, &out); err != nil {
		return err
	}
	if jsonOut {
		_, err := fmt.Fprintln(deps.Stdout, string(out))
		return err
	}
	parts := []string{}
	for _, name := range req.Add {
		parts = append(parts, "+"+name)
	}
	for _, name := range req.Remove {
		parts = append(parts, "-"+name)
	}
	_, err = fmt.Fprintf(deps.Stdout, "%s %d: %s\n", kind, id, strings.Join(parts, " "))
	return err
}

func printTags(w io.Writer, tags []sharedtypes.Tag) error {
	if len(tags) == 0 {
		_, err := fmt.Fprintln(w, "no tags")
		return err
	}
	width := 4
	for _, tag := range tags {
		width = max(width, len(tag.Name)+1)
	}
	if _, err := fmt.Fprintf(w, "%-*s  %6s  %6s  %-8s  %s\n", width, "TAG", "ISSUES", "HABITS", "COLOR", "KIND"); err != nil {
		return err
	}
	for _, tag := range tags {
		color := "-"
		if tag.Color != nil && *tag.Color != "" {
			color = *tag.Color
		}
		kind := "user"
		if tag.System {
			kind = "system"
		}
		if _, err := fmt.Fprintf(w, "%-*s  %6d  %6d  %-8s  %s\n", width, "+"+tag.Name, tag.IssueCount, tag.HabitCount, color, kind); err != nil {
			return err
		}
	}
	return nil
}

func joinTags(names []string) string {
	out := make([]string, len(names))
	for i, name := range names {
		out[i] = "+" + name
	}
	return strings.Join(out, " ")
}
