package error_handling

import (
	"encoding/gob"
	"log"
	"os"
)

type subprocessErrorReceiver struct {
	fifo    *os.File
	decoder *gob.Decoder
}

func newSubprocessErrorReceiver(fifoName string) (*subprocessErrorReceiver, error) {
	fifo, err := os.OpenFile(fifoName, os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		return nil, err
	}
	decoder := gob.NewDecoder(fifo)
	receiver := subprocessErrorReceiver{fifo, decoder}
	return &receiver, nil
}

func (r *subprocessErrorReceiver) receive() error {
	var wrapper subprocessErrorWrapper
	err := r.decoder.Decode(&wrapper)
	if err != nil {
		log.Panicf("Unable to retrieve daemon status\n%v", err.Error())
	}
	return wrapper.unwrap()
}

func (r *subprocessErrorReceiver) Close() {
	_ = r.fifo.Close()
}
