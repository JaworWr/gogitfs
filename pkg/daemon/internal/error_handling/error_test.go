package error_handling

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"syscall"
	"testing"
)

func mkFifo(t *testing.T) string {
	tmpdir := t.TempDir()
	namedPipe := filepath.Join(tmpdir, "aaa.pipe")
	err := syscall.Mkfifo(namedPipe, 0700)
	if err != nil {
		t.Fatalf("Cannot create named pipe %v. Error: %v", namedPipe, err)
	}

	return namedPipe
}

func sendMsg(t *testing.T, namedPipe string, val error) {
	sender, err := NewSubprocessErrorSender(namedPipe)
	if err != nil {
		t.Fatalf("Cannot create sender. Error: %v", err)
	}
	wrapped := wrapError(val)
	sender.send(wrapped)
	sender.Close()
}

func recvMsg(t *testing.T, namedPipe string) (val error) {
	receiver, err := NewSubprocessErrorReceiver(namedPipe)
	if err != nil {
		t.Fatalf("Cannot create receiver. Error: %v", err)
	}
	val = receiver.Receive()
	receiver.Close()
	return
}

func Test_Send_Receive(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{"nil", nil},
		{"error", errors.New("aaa")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			namedPipe := mkFifo(t)
			defer func() { _ = syscall.Unlink(namedPipe) }()

			go sendMsg(t, namedPipe, tc.err)
			received := recvMsg(t, namedPipe)
			if tc.err == nil {
				assert.Nil(t, received, "expected a nil error")
			} else {
				assert.NotNil(t, received, "expected a non-nil error")
				assert.Equal(t, tc.err.Error(), received.Error(), "sent end received error messages don't match")
			}
		})
	}
}
