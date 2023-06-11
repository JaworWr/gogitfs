package error_handler

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

type fatalHandlerType struct{}

func (h fatalHandlerType) HandleError(err error) {
	log.Fatalf("[ERROR] An error occurred: %v", err.Error())
}

var FatalHandler fatalHandlerType

type LogHandlerWrapper struct {
	next ErrorHandler
}

func (h *LogHandlerWrapper) HandleError(err error) {
	log.Printf("An error occurred: %v", err.Error())
	h.next.HandleError(err)
}

var LogHandler = &LogHandlerWrapper{NoOpHandler}
