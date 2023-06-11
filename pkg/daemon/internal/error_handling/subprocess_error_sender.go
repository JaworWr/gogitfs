package error_handling

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
)

type SubprocessErrorSender struct {
	fifo    *os.File
	encoder *gob.Encoder
}

func NewSubprocessErrorSender() (*SubprocessErrorSender, error) {
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
	sender := SubprocessErrorSender{fifo, encoder}
	return &sender, nil
}

func (s *SubprocessErrorSender) send(wrapper *subprocessErrorWrapper) {
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
	err = s.fifo.Sync()
	if err != nil {
		return
	}
	_ = s.fifo.Close()
	return
}
