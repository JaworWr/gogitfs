package error_handling

import (
	"encoding/gob"
	"log"
	"os"
)

type SubprocessErrorSender struct {
	errorSent bool
	fifo      *os.File
	encoder   *gob.Encoder
}

func NewSubprocessErrorSender(fifoName string) (*SubprocessErrorSender, error) {
	fifo, err := os.OpenFile(fifoName, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		return nil, err
	}
	encoder := gob.NewEncoder(fifo)
	sender := SubprocessErrorSender{fifo: fifo, encoder: encoder}
	return &sender, nil
}

func (s *SubprocessErrorSender) send(wrapper *subprocessErrorWrapper) {
	s.errorSent = true
	encodeErr := s.encoder.Encode(wrapper)
	if encodeErr != nil {
		log.Panicf("Daemon status send error\n%v", encodeErr.Error())
	}
}

func (s *SubprocessErrorSender) HandleError(err error) {
	wrapper := wrapError(err)
	s.send(wrapper)
}

func (s *SubprocessErrorSender) HandleSuccess() {
	wrapper := wrapError(nil)
	s.send(wrapper)
}

func (s *SubprocessErrorSender) Close() (err error) {
	if !s.errorSent {
		s.HandleError(&UnknownError{})
	}
	err = s.fifo.Sync()
	if err != nil {
		return
	}
	_ = s.fifo.Close()
	return
}
