package daemon

import (
	"flag"
	"fmt"
	"os"
	"strconv"
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
	Expected int
	Got      int
}

func (err *NotEnoughArgsError) Error() string {
	return fmt.Sprintf("not enough positional arguments: expected %v, got %v", err.Expected, err.Got)
}

type TooManyArgsError struct {
	ExtraArgs []string
}

func (err *TooManyArgsError) Error() string {
	var quoted []string
	for _, arg := range err.ExtraArgs {
		quoted = append(quoted, "\""+arg+"\"")
	}
	unexpected := strings.Join(quoted, ", ")
	return "unexpected arguments: " + unexpected
}

func SetupFlags(da DaemonArgs) {
	da.Setup()
	flag.Usage = func() {
		var argnames string
		for _, arg := range da.PositionalArgs() {
			argnames += " <" + arg.Name + ">"
		}
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s%s\n", os.Args[0], argnames)
		for _, arg := range da.PositionalArgs() {
			_, _ = fmt.Fprintf(flag.CommandLine.Output(), "  %s\t%s\n", arg.Name, arg.Usage)
		}
		flag.PrintDefaults()
	}
}

func argsToFullList(da DaemonArgs) []string {
	result := []string{os.Args[0]}
	result = append(result, da.Serialize()...)
	return result
}

func SerializeStringFlag(flag string, value string) string {
	return fmt.Sprintf("--%s=%s", flag, value)
}

func SerializeBoolFlag(flag string, value bool) string {
	var valueStr string
	if value {
		valueStr = "true"
	} else {
		valueStr = "false"
	}
	return SerializeStringFlag(flag, valueStr)
}

func SerializeIntFlag(flag string, value int64) string {
	valueStr := strconv.FormatInt(value, 10)
	return SerializeStringFlag(flag, valueStr)
}

func SerializeUintFlag(flag string, value uint64) string {
	valueStr := strconv.FormatUint(value, 10)
	return SerializeStringFlag(flag, valueStr)
}
