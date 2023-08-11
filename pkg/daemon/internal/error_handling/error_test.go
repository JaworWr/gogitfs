package error_handling

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"syscall"
	"testing"
)

func getPair(t *testing.T) (string, *SubprocessErrorSender, *SubprocessErrorReceiver) {
	tmpdir := t.TempDir()
	namedPipe := filepath.Join(tmpdir, "aaa.pipe")
	err := syscall.Mkfifo(namedPipe, 0700)
	if err != nil {
		t.Fatalf("Cannot create named pipe %v. Error: %v", namedPipe, err)
	}

	sender, err := NewSubprocessErrorSender(namedPipe)
	if err != nil {
		t.Fatalf("Cannot create sender. Error: %v", err)
	}
	receiver, err := NewSubprocessErrorReceiver(namedPipe)
	if err != nil {
		t.Fatalf("Cannot create receiver. Error: %v", err)
	}
	return namedPipe, sender, receiver
}

func Test_Send_Receive(t *testing.T) {
	// TODO - doesn't work; try with goroutines
	testCases := []struct {
		name string
		err  error
	}{
		{"nil", nil},
		{"error", errors.New("aaa")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			namedPipe, sender, receiver := getPair(t)
			defer func() { _ = syscall.Unlink(namedPipe) }()

			wrapped := wrapError(tc.err)
			sender.send(wrapped)
			sender.Close()
			received := receiver.Receive()
			assert.Equal(t, tc.err, received)
		})
	}
}
