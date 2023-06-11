package error_handling

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
)

type subprocessErrorSender struct {
	fifo    *os.File
	encoder *gob.Encoder
}

func newSubprocessErrorSender() (*subprocessErrorSender, error) {
	fifoName, ok := os.LookupEnv(pipeKey)
	if !ok {
		err := fmt.Errorf("missing environment variable: %v", pipeKey)
		return nil, err
	}
	fifo, err := os.OpenFile(fifoName, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		return nil, err
	}
	encoder := gob.NewEncoder(fifo)
	sender := subprocessErrorSender{fifo, encoder}
	return &sender, nil
}

func (s *subprocessErrorSender) send(wrapper *subprocessErrorWrapper) {
	encodeErr := s.encoder.Encode(wrapper)
	if encodeErr != nil {
		log.Panicf("Send error\n%v", encodeErr.Error())
	}
}

func (s *subprocessErrorSender) HandleError(err error) {
	wrapper := wrapError(err)
	s.send(wrapper)
}

func (s *subprocessErrorSender) HandleSuccess() {
	wrapper := wrapError(nil)
	s.send(wrapper)
}

func (s *subprocessErrorSender) Close() (err error) {
	err = s.fifo.Sync()
	if err != nil {
		return
	}
	_ = s.fifo.Close()
	return
}
