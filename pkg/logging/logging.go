package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

type LogLevelFlag int

const (
	Debug LogLevelFlag = iota
	Info
	Warning
	Error
)

var levelToStr = map[LogLevelFlag]string{
	Debug:   "DEBUG",
	Info:    "INFO",
	Warning: "WARNING",
	Error:   "ERROR",
}

var strToLevel = map[string]LogLevelFlag{
	"DEBUG":   Debug,
	"INFO":    Info,
	"WARNING": Warning,
	"ERROR":   Error,
}

func (f *LogLevelFlag) String() string {
	return levelToStr[*f]
}

func (f *LogLevelFlag) Set(s string) error {
	val, ok := strToLevel[s]
	if ok {
		*f = val
		return nil
	}
	val1, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*f = LogLevelFlag(val1)
	return nil
}

var logLevel = Info

var DebugLog *log.Logger
var InfoLog *log.Logger
var WarningLog *log.Logger
var ErrorLog *log.Logger

func LoggerWithLevel(l LogLevelFlag) *log.Logger {
	switch l {
	case Debug:
		return DebugLog
	case Info:
		return InfoLog
	case Warning:
		return WarningLog
	case Error:
		return ErrorLog
	}
	return nil
}

func Init(l LogLevelFlag) {
	logLevel = l
	DebugLog = makeLogger(Debug)
	InfoLog = makeLogger(Info)
	WarningLog = makeLogger(Warning)
	ErrorLog = makeLogger(Error)
}

func makeLogger(level LogLevelFlag) *log.Logger {
	var output io.Writer
	if level >= logLevel {
		output = os.Stdout
	} else {
		output = io.Discard
	}

	prefix := fmt.Sprintf("[%s] ", levelToStr[level])
	return log.New(output, prefix, log.LstdFlags|log.Lmsgprefix)
}
