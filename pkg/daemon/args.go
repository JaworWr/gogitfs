package daemon

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// PositionalArg represents a positional command line argument.
// This struct is used to generate better help messages.
type PositionalArg struct {
	Name  string
	Usage string
}

// DaemonArgs manages command line arguments required by the daemon.
// This interface should be implemented by structs describing daemon options.
type DaemonArgs interface {
	// Setup sets up parsing of command line flags.
	Setup()
	// PositionalArgs returns an array of expected positional arguments.
	PositionalArgs() []PositionalArg
	// HandlePositionalArgs parses positional arguments provided on the command line
	HandlePositionalArgs([]string) error
	// Serialize converts current values into parseable command line arguments.
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

// SetupFlags sets up command line flags and usage string for the given DaemonArgs object.
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
	// prepend process name to arguments
	result := []string{os.Args[0]}
	result = append(result, da.Serialize()...)
	return result
}

// helper functions for flag serialization

func SerializeStringFlag(flag string, value string) string {
	return fmt.Sprintf("--%s=%s", flag, value)
}

var _ = SerializeStringFlag

func SerializeBoolFlag(flag string, value bool) string {
	var valueStr string
	if value {
		valueStr = "true"
	} else {
		valueStr = "false"
	}
	return SerializeStringFlag(flag, valueStr)
}

var _ = SerializeBoolFlag

func SerializeIntFlag(flag string, value int64) string {
	valueStr := strconv.FormatInt(value, 10)
	return SerializeStringFlag(flag, valueStr)
}

var _ = SerializeIntFlag

func SerializeUintFlag(flag string, value uint64) string {
	valueStr := strconv.FormatUint(value, 10)
	return SerializeStringFlag(flag, valueStr)
}

var _ = SerializeUintFlag
