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
	if s.errorSent {
		log.Println("Attempting to send multiple errors!")
		return
	}
	s.errorSent = true
	err := s.encoder.Encode(wrapper)
	if err != nil {
		log.Panicf("Daemon status send error\n%v", err.Error())
	}
}

func (s *SubprocessErrorSender) HandleError(err error) {
	wrapper := wrapError(err)
	s.send(wrapper)
	os.Exit(1)
}

func (s *SubprocessErrorSender) HandleSuccess() {
	wrapper := wrapError(nil)
	s.send(wrapper)
}

func (s *SubprocessErrorSender) Close() {
	if !s.errorSent {
		s.HandleError(&UnknownError{})
	}
	_ = s.fifo.Close()
	return
}
