package error_handler

import (
	"gogitfs/pkg/logging"
	"log"
)

type ErrorHandler interface {
	HandleError(err error)
}

type LogHandlerWrapper struct {
	next   ErrorHandler
	logger *log.Logger
}

func (h *LogHandlerWrapper) HandleError(err error) {
	funcname := logging.CurrentFuncName(1, logging.Full)
	logging.WarningLog.Printf("[%v] %v", funcname, err.Error())
	h.next.HandleError(err)
}

func MakeLoggingHandler(h ErrorHandler, level logging.LogLevelFlag) *LogHandlerWrapper {
	return &LogHandlerWrapper{next: h, logger: logging.LoggerWithLevel(level)}
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
