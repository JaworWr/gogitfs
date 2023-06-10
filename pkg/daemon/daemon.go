package daemon

import (
	"gogitfs/pkg/error_handler"
)

type DaemonProcess interface {
	DaemonArgs(args []string) []string
	DaemonProcess(handler error_handler.ErrorHandler)
}
