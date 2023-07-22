package daemon

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type DaemonArgs interface {
	SetupFlags()
	PositionalArgNames() []string
	HandlePositionalArgs([]string) error
}

type NotEnoughArgsError struct {
	expected int
	got      int
}

func (err *NotEnoughArgsError) Error() string {
	return fmt.Sprintf("not enough arguments: expected %v, got %v", err.expected, err.got)
}

type TooManyArgsError struct {
	extraArgs []string
}

func (err *TooManyArgsError) Error() string {
	var quoted []string
	for _, arg := range err.extraArgs {
		quoted = append(quoted, "\""+arg+"\"")
	}
	unexpected := strings.Join(quoted, ", ")
	return "unexpected arguments: " + unexpected
}

func InitArgs(da DaemonArgs) {
	da.SetupFlags()
	flag.Usage = func() {
		var argnames string
		for _, argname := range da.PositionalArgNames() {
			argnames += " " + argname
		}
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s%s\n", os.Args[0], argnames)
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s", os.Args[0])
		flag.PrintDefaults()
	}
}
