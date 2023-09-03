package error_handling

import (
	"encoding/gob"
	"os"
)

// SubprocessErrorReceiver receives the error or success notification from the daemon.
type SubprocessErrorReceiver struct {
	fifo    *os.File
	decoder *gob.Decoder
}

func NewSubprocessErrorReceiver(namedPipeName string) (*SubprocessErrorReceiver, error) {
	fifo, err := os.OpenFile(namedPipeName, os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		return nil, err
	}
	decoder := gob.NewDecoder(fifo)
	receiver := SubprocessErrorReceiver{fifo, decoder}
	return &receiver, nil
}

// Receive waits until status information is available from the daemon. Received error, if any, is returned.
func (r *SubprocessErrorReceiver) Receive() error {
	var wrapper subprocessErrorWrapper
	err := r.decoder.Decode(&wrapper)
	if err != nil {
		panic("Unable to retrieve daemon status\nError: " + err.Error())
	}
	return wrapper.unwrap()
}

func (r *SubprocessErrorReceiver) Close() {
	_ = r.fifo.Close()
}
