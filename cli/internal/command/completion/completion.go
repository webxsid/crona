package completion

import (
	"fmt"
	"io"
	"strings"
)

func Usage() string {
	return "Usage: crona completion <zsh|bash|fish>\n"
}

func Run(args []string, cliName string, stdout io.Writer) error {
	if len(args) == 0 || (len(args) == 1 && isHelpArg(args[0])) {
		_, err := fmt.Fprint(stdout, Usage())
		return err
	}
	if len(args) != 1 {
		return fmt.Errorf("usage: %s", strings.TrimSpace(Usage()))
	}
	switch args[0] {
	case "zsh":
		_, err := fmt.Fprint(stdout, zsh(cliName))
		return err
	case "bash":
		_, err := fmt.Fprint(stdout, bash(cliName))
		return err
	case "fish":
		_, err := fmt.Fprint(stdout, fish(cliName))
		return err
	default:
		return fmt.Errorf("unknown shell: %s", args[0])
	}
}

func isHelpArg(value string) bool {
	switch strings.TrimSpace(value) {
	case "-h", "--help":
		return true
	default:
		return false
	}
}

func zsh(name string) string {
	return fmt.Sprintf(`#compdef %s
_%s() {
  local -a commands
  commands=('kernel:Kernel commands' 'completion:Shell completions' 'context:Context commands' 'timer:Timer commands' 'issue:Issue commands' 'update:Update commands' 'export:Export commands' 'dev:Dev-only commands')
  if (( CURRENT == 2 )); then
    _describe 'command' commands
    return
  fi
  case "${words[2]}" in
    kernel) _values 'kernel command' attach detach restart wipe-data info status ;;
    completion) _values 'shell' zsh bash fish ;;
    context) _values 'context command' get set clear clear-issue switch-repo switch-stream switch-issue ;;
    timer) _values 'timer command' status start pause resume end ;;
    issue) _values 'issue command' start ;;
    update) _values 'update command' status check dismiss notes ;;
    export) _values 'export command' daily weekly repo stream issue-rollup csv calendar reports ;;
    dev) _values 'dev command' seed clear ;;
  esac
}
_%s "$@"
`, name, name, name)
}

func bash(name string) string {
	return fmt.Sprintf(`_%s()
{
  local cur prev words cword
  _init_completion || return
  if [[ ${cword} -eq 1 ]]; then
    COMPREPLY=( $(compgen -W "kernel completion context timer issue update export dev" -- "$cur") )
    return
  fi
  case "${words[1]}" in
    kernel) COMPREPLY=( $(compgen -W "attach detach restart wipe-data info status" -- "$cur") ) ;;
    completion) COMPREPLY=( $(compgen -W "zsh bash fish" -- "$cur") ) ;;
    context) COMPREPLY=( $(compgen -W "get set clear clear-issue switch-repo switch-stream switch-issue" -- "$cur") ) ;;
    timer) COMPREPLY=( $(compgen -W "status start pause resume end" -- "$cur") ) ;;
    issue) COMPREPLY=( $(compgen -W "start" -- "$cur") ) ;;
    update) COMPREPLY=( $(compgen -W "status check dismiss notes" -- "$cur") ) ;;
    export) COMPREPLY=( $(compgen -W "daily weekly repo stream issue-rollup csv calendar reports" -- "$cur") ) ;;
    dev) COMPREPLY=( $(compgen -W "seed clear" -- "$cur") ) ;;
  esac
}
complete -F _%s %s
`, name, name, name)
}

func fish(name string) string {
	return fmt.Sprintf(`complete -c %s -f -n "__fish_use_subcommand" -a "kernel completion context timer issue update export dev"
complete -c %s -f -n "__fish_seen_subcommand_from kernel" -a "attach detach restart wipe-data info status"
complete -c %s -f -n "__fish_seen_subcommand_from completion" -a "zsh bash fish"
complete -c %s -f -n "__fish_seen_subcommand_from context" -a "get set clear clear-issue switch-repo switch-stream switch-issue"
complete -c %s -f -n "__fish_seen_subcommand_from timer" -a "status start pause resume end"
complete -c %s -f -n "__fish_seen_subcommand_from issue" -a "start"
complete -c %s -f -n "__fish_seen_subcommand_from update" -a "status check dismiss notes"
complete -c %s -f -n "__fish_seen_subcommand_from export" -a "daily weekly repo stream issue-rollup csv calendar reports"
complete -c %s -f -n "__fish_seen_subcommand_from dev" -a "seed clear"
`, name, name, name, name, name, name, name, name, name)
}
