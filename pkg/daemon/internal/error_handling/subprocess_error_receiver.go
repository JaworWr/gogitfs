package error_handling

import (
	"encoding/gob"
	"log"
	"os"
)

type SubprocessErrorReceiver struct {
	fifo    *os.File
	decoder *gob.Decoder
}

func NewSubprocessErrorReceiver(fifoName string) (*SubprocessErrorReceiver, error) {
	fifo, err := os.OpenFile(fifoName, os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		return nil, err
	}
	decoder := gob.NewDecoder(fifo)
	receiver := SubprocessErrorReceiver{fifo, decoder}
	return &receiver, nil
}

func (r *SubprocessErrorReceiver) Receive() error {
	var wrapper subprocessErrorWrapper
	err := r.decoder.Decode(&wrapper)
	if err != nil {
		log.Panicf("Unable to retrieve daemon status\n%v", err.Error())
	}
	return wrapper.unwrap()
}

func (r *SubprocessErrorReceiver) Close() {
	_ = r.fifo.Close()
}
