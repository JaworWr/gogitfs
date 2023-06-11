package error_handling

import (
	"encoding/gob"
	"os"
)

type subprocessErrorSender struct {
	fifo    *os.File
	encoder gob.Encoder
}

func (s *subprocessErrorSender) HandleError(err error) {
	//TODO implement me
	panic("implement me")
}

func (s *subprocessErrorSender) HandleSuccess() {
	//TODO implement me
	panic("implement me")
}
