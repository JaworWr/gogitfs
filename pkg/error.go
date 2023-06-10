package pkg

import (
	"log"
)

type ErrorHandler interface {
	HandleError(err error)
}

type noOpHandlerType struct{}

func (h noOpHandlerType) HandleError(_ error) {

}

var NoOpHandler noOpHandlerType

type panicHandlerType struct{}

func (h panicHandlerType) HandleError(err error) {
	log.Fatalf("[ERROR] An error occurred: %v", err.Error())
}

var PanicHandler panicHandlerType

type LogHandlerWrapper struct {
	next ErrorHandler
}

func (h *LogHandlerWrapper) HandleError(err error) {
	log.Printf("An error occurred: %v", err.Error())
	h.next.HandleError(err)
}

var LogHandler = &LogHandlerWrapper{NoOpHandler}
