package flags

import (
	"flag"
	"strings"
)

type discard struct{}

func (discard) Write(p []byte) (int, error) {
	return len(p), nil
}

func New(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(discard{})
	return fs
}

func HasJSON(args []string) bool {
	for _, arg := range args {
		if arg == "--json" {
			return true
		}
	}
	return false
}

func IsHelpArg(value string) bool {
	switch strings.TrimSpace(value) {
	case "-h", "--help":
		return true
	default:
		return false
	}
}

// ParseWithPositionals lets flags appear before or after the positional names.
func ParseWithPositionals(fs interface {
	Parse([]string) error
	Args() []string
}, args []string) ([]string, error) {
	positionals := []string{}
	for len(args) > 0 {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		rest := fs.Args()
		if len(rest) == 0 {
			break
		}
		positionals = append(positionals, rest[0])
		args = rest[1:]
	}
	return positionals, nil
}
