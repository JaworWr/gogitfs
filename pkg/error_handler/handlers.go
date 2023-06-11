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
	log.Printf("[ERROR] An error occurred: %v", err.Error())
	h.next.HandleError(err)
}

func MakeLoggingHandler(h ErrorHandler) *LogHandlerWrapper {
	return &LogHandlerWrapper{next: h}
}
