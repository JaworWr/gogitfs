package error_handler

import (
	"gogitfs/pkg/logging"
)

type ErrorHandler interface {
	HandleError(err error)
}

type LogHandlerWrapper struct {
	next ErrorHandler
}

func (h *LogHandlerWrapper) HandleError(err error) {
	funcname := logging.CurrentFuncName(1, logging.Full)
	logging.WarningLog.Printf("[%v] %v", funcname, err.Error())
	h.next.HandleError(err)
}

func MakeLoggingHandler(h ErrorHandler) *LogHandlerWrapper {
	return &LogHandlerWrapper{next: h}
}

type noOpErrorHandler struct{}

func (h noOpErrorHandler) HandleError(_ error) {

}

var Logging = MakeLoggingHandler(noOpErrorHandler{})

type fatalErrorHandler struct{}

func (h fatalErrorHandler) HandleError(err error) {
	funcname := logging.CurrentFuncName(1, logging.Full)
	logging.ErrorLog.Printf("[%v] %v", funcname, err.Error())
}

var Fatal fatalErrorHandler
