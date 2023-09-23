package daemon

import (
	"flag"
	"fmt"
	"github.com/sevlyar/go-daemon"
	"gogitfs/pkg/daemon/internal/environment"
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

// CliArgs manages command line arguments required by the daemon.
// This interface should be implemented by structs describing daemon options.
type CliArgs interface {
	// Setup sets up parsing of command line flags.
	Setup()
	// PositionalArgs returns an array of expected positional arguments.
	PositionalArgs() []PositionalArg
	// HandlePositionalArgs parses positional arguments provided on the command line
	HandlePositionalArgs([]string) error
}

type SerializableCliArgs interface {
	CliArgs
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

// ParseFlags sets up command line flags and usage string for the given CliArgs object.
// parentSetup is meant to set up flags
func ParseFlags(ca CliArgs, parentSetup func()) error {
	environment.SetupFlags()
	ca.Setup()
	if !daemon.WasReborn() {
		parentSetup()
	}
	flag.Usage = func() {
		var argnames string
		for _, arg := range ca.PositionalArgs() {
			argnames += " <" + arg.Name + ">"
		}
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s%s\n", os.Args[0], argnames)
		for _, arg := range ca.PositionalArgs() {
			_, _ = fmt.Fprintf(flag.CommandLine.Output(), "  %s\t%s\n", arg.Name, arg.Usage)
		}
		flag.PrintDefaults()
	}
	flag.Parse()
	err := ca.HandlePositionalArgs(flag.Args())
	return err
}

func argsToFullList(ca SerializableCliArgs) []string {
	// prepend process name to arguments
	result := []string{os.Args[0]}
	result = append(result, ca.Serialize()...)
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
