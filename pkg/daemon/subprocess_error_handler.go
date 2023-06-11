package daemon

import "gogitfs/pkg/error_handler"

type subprocessErrorHandler struct {
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
var _ SuccessHandler = (*subprocessErrorHandler)(nil)
