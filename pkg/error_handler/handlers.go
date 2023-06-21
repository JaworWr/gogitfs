package error_handler

import (
	"log"
)

type ErrorHandler interface {
	HandleError(err error)
}

type LogHandlerWrapper struct {
	next ErrorHandler
}

func (h *LogHandlerWrapper) HandleError(err error) {
	log.Printf("[ERROR] %v", err.Error())
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
	log.Fatalf("[ERROR] %v", err.Error())
}

var Fatal fatalErrorHandler
