// Package error_handler contains utilities for handling errors.
package error_handler

import (
	"gogitfs/pkg/logging"
)

type ErrorHandler interface {
	HandleError(err error)
}

// LogHandlerWrapper wraps another error handler emitting a message with a specified log level.
type LogHandlerWrapper struct {
	Next     ErrorHandler
	LogLevel logging.LogLevelFlag
}

func (h *LogHandlerWrapper) HandleError(err error) {
	funcname := logging.CurrentFuncName(1, logging.Full)
	logger := logging.LoggerWithLevel(h.LogLevel)
	logger.Printf("[%v] %v", funcname, err.Error())
	h.Next.HandleError(err)
}

// MakeLoggingHandler is a helper function for logger wrapping.
func MakeLoggingHandler(h ErrorHandler, level logging.LogLevelFlag) *LogHandlerWrapper {
	return &LogHandlerWrapper{Next: h, LogLevel: level}
}

type noOpErrorHandler struct{}

func (h noOpErrorHandler) HandleError(_ error) {

}

// Logging handler simply displays a message with level "ERROR"
var Logging = MakeLoggingHandler(noOpErrorHandler{}, logging.Error)

type fatalErrorHandler struct{}

func (h fatalErrorHandler) HandleError(err error) {
	funcname := logging.CurrentFuncName(1, logging.Full)
	logging.ErrorLog.Fatalf("[%v] %v", funcname, err.Error())
}

// Fatal handler displays a message with level "ERROR" and terminates the program.
var Fatal fatalErrorHandler
