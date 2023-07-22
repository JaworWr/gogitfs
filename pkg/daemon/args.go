package daemon

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type PositionalArg struct {
	Name  string
	Usage string
}

type DaemonArgs interface {
	Setup()
	PositionalArgs() []PositionalArg
	HandlePositionalArgs([]string) error
	Serialize() []string
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
	da.Setup()
	flag.Usage = func() {
		var argnames string
		for _, arg := range da.PositionalArgs() {
			argnames += " <" + arg.Name + ">"
		}
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s%s\n", os.Args[0], argnames)
		for _, arg := range da.PositionalArgs() {
			_, _ = fmt.Fprintf(flag.CommandLine.Output(), "  %s\t%s", arg.Name, arg.Usage)
		}
		flag.PrintDefaults()
	}
}
