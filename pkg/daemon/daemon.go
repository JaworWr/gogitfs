package daemon

import (
	"gogitfs/pkg/error_handler"
)

type DaemonProcess interface {
	DaemonArgs(args []string) []string
	DaemonProcess(errHandler error_handler.ErrorHandler, succHandler SuccessHandler)
}
