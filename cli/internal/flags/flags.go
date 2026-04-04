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
