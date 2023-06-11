package error_handling

import (
	"encoding/gob"
	"os"
)

type subprocessErrorReceiver struct {
	fifo    *os.File
	decoder gob.Decoder
}
