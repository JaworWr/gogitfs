package logging

import (
	"fmt"
	"io"
	"log"
	"os"
)

type LogLevelFlag int

const (
	Debug LogLevelFlag = iota
	Info
	Warning
	Error
)

var logLevel = Info

var DebugLog *log.Logger
var InfoLog *log.Logger
var WarningLog *log.Logger
var ErrorLog *log.Logger

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

	var prefix string
	switch level {
	case Debug:
		prefix = "DEBUG"
	case Info:
		prefix = "INFO"
	case Warning:
		prefix = "WARNING"
	case Error:
		prefix = "ERROR"
	}
	prefix = fmt.Sprintf("[%s] ", prefix)

	return log.New(output, prefix, log.LstdFlags|log.Lmsgprefix)
}
