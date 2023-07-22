package error_handler

import (
	"gogitfs/pkg/logging"
)

type ErrorHandler interface {
	HandleError(err error)
}

type LogHandlerWrapper struct {
	next     ErrorHandler
	logLevel logging.LogLevelFlag
}

func (h *LogHandlerWrapper) HandleError(err error) {
	funcname := logging.CurrentFuncName(1, logging.Full)
	logger := logging.LoggerWithLevel(h.logLevel)
	logger.Printf("[%v] %v", funcname, err.Error())
	h.next.HandleError(err)
}

func MakeLoggingHandler(h ErrorHandler, level logging.LogLevelFlag) *LogHandlerWrapper {
	return &LogHandlerWrapper{next: h, logLevel: level}
}

type noOpErrorHandler struct{}

func (h noOpErrorHandler) HandleError(_ error) {

}

var Logging = MakeLoggingHandler(noOpErrorHandler{}, logging.Warning)

type fatalErrorHandler struct{}

func (h fatalErrorHandler) HandleError(err error) {
	funcname := logging.CurrentFuncName(1, logging.Full)
	logging.ErrorLog.Printf("[%v] %v", funcname, err.Error())
}

var Fatal fatalErrorHandler
