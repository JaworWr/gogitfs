// Package logging provides utilities for log formatting and level-based filtering.
package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

// LogLevelFlag represents log levels, with 0 (DEBUG) being the most detailed
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

// String returns a string representation of a log level flag
func (f *LogLevelFlag) String() string {
	return levelToStr[*f]
}

// Set parses log level flag, either from an integer or an all-caps name.
func (f *LogLevelFlag) Set(s string) error {
	val, ok := strToLevel[s]
	if ok {
		*f = val
		return nil
	}
	val1, err := strconv.Atoi(s)
	val = LogLevelFlag(val1)
	if val < 0 || val > Error {
		err = fmt.Errorf("log level must be between 0 and %v, got %v", Error, val)
	}
	if err != nil {
		return err
	}
	*f = val
	return nil
}

// current logging level
var logLevel = Info

// DebugLog writes messages with level DEBUG
var DebugLog *log.Logger

// InfoLog writes messages with level INFO
var InfoLog *log.Logger

// WarningLog writes messages with level WARNING
var WarningLog *log.Logger

// ErrorLog writes messages with level ERROR
var ErrorLog *log.Logger

// LoggerWithLevel returns the logger for the specified level
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

// MakeFileLogger returns a logger writing to the specified file
func MakeFileLogger(name string) (*log.Logger, error) {
	file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	logger := log.New(file, "", log.LstdFlags)
	return logger, nil
}
