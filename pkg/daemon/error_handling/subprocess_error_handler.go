package error_handling

import (
	"gogitfs/pkg/daemon"
	"gogitfs/pkg/error_handler"
)

type subprocessErrorHandler struct {
}

type subprocessErrorReceiver struct {
}

func (s subprocessErrorHandler) HandleError(err error) {
	//TODO implement me
	panic("implement me")
}

func (s subprocessErrorHandler) HandleSuccess() {
	//TODO implement me
	panic("implement me")
}

var _ error_handler.ErrorHandler = (*subprocessErrorHandler)(nil)
var _ daemon.SuccessHandler = (*subprocessErrorHandler)(nil)
