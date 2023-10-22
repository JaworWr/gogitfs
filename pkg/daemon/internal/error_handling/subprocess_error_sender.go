package error_handling

import (
	"encoding/gob"
	"fmt"
	"os"
)

// SubprocessErrorSender sends an error or success notification from the daemon to the parent process.
type SubprocessErrorSender struct {
	errorSent bool
	fifo      *os.File
	encoder   *gob.Encoder
}

func NewSubprocessErrorSender(namedPipeName string) (*SubprocessErrorSender, error) {
	fifo, err := os.OpenFile(namedPipeName, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		return nil, fmt.Errorf("cannot open named pipe for writing: %w", err)
	}
	encoder := gob.NewEncoder(fifo)
	sender := SubprocessErrorSender{fifo: fifo, encoder: encoder}
	return &sender, nil
}

// send sends the wrapped value to the parent process
func (s *SubprocessErrorSender) send(wrapper *subprocessErrorWrapper) {
	if s.errorSent {
		panic("Attempted to send multiple errors")
	}
	s.errorSent = true
	err := s.encoder.Encode(wrapper)
	if err != nil {
		panic("Failed to send daemon status\nError: " + err.Error())
	}
}

// HandleError sends the error to the parent process. Afterward, the daemon process exits immediately.
// This method should not be called after HandleSuccess.
func (s *SubprocessErrorSender) HandleError(err error) {
	wrapper := wrapError(err)
	s.send(wrapper)
	os.Exit(1)
}

// HandleSuccess notifies the parent process about successful daemon startup. It should only be called once.
func (s *SubprocessErrorSender) HandleSuccess() {
	wrapper := wrapError(nil)
	s.send(wrapper)
}

func (s *SubprocessErrorSender) Close() {
	if !s.errorSent {
		s.HandleError(UnknownError)
	}
	_ = s.fifo.Close()
	return
}
