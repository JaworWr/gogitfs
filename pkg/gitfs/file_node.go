package gitfs

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/error_handler"
	"gogitfs/pkg/gitfs/internal/utils"
	"gogitfs/pkg/logging"
	"io"
	"syscall"
)

// fileNode represents a regular file in a Git tree
type fileNode struct {
	fs.Inode
	file *object.File
	attr fuse.Attr
	// data contains the actual file data, lazily read during opening
	data []byte
}

// newFileNode creates a new fileNode based on File object and attributes extracted from commit
func newFileNode(file *object.File, attr fuse.Attr) *fileNode {
	node := &fileNode{}
	node.file = file
	node.attr = attr
	node.attr.Mode = 0444
	node.attr.Size = uint64(node.file.Size)
	return node
}

// fileNodeHandle is a dummy handle to be returned when opening the file
type fileNodeHandle struct{}

func (n *fileNode) GetCallCtx() logging.CallCtx {
	info := utils.NodeCallCtx(n)
	info["name"] = n.file.Name
	info["mode"] = n.file.Mode
	return info
}

// Getattr returns attributes corresponding to the tree of the given commit (passed during newFileNode call).
func (n *fileNode) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	logging.LogCall(n, nil)
	out.Attr = n.attr
	return fs.OK
}

// Open prepares the node to be read and returnes a dummy handle. If the data array has not been initialized yet,
// it will be populated by data read from the Git blob.
func (n *fileNode) Open(_ context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	logging.LogCall(n, logging.CallCtx{"flags": flags})
	if flags&(syscall.O_RDWR|syscall.O_WRONLY) != 0 {
		return nil, 0, syscall.EROFS
	}

	if n.data == nil {
		reader, err := n.file.Reader()
		if err != nil {
			error_handler.Logging.HandleError(
				fmt.Errorf("cannot create reader: %w", err),
			)
			return nil, 0, syscall.EIO
		}
		defer func(reader io.ReadCloser) {
			_ = reader.Close()
		}(reader)
		buf := &bytes.Buffer{}
		_, err = buf.ReadFrom(reader)
		if err != nil {
			error_handler.Logging.HandleError(
				fmt.Errorf("cannot perform ReadFrom: %w", err),
			)
			return nil, 0, syscall.EIO
		}
		n.data = buf.Bytes()
	}
	return fileNodeHandle{}, 0, fs.OK
}

// Read returns the data of the blob object corresponding to the desired file.
func (n *fileNode) Read(_ context.Context, f fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	logging.LogCall(n, logging.CallCtx{"max": len(dest), "off": off})
	if f == nil {
		return nil, syscall.EIO
	}

	toRead := min(int64(len(dest)), int64(len(n.data))-off)
	copy(dest, n.data[off:off+toRead])
	res := fuse.ReadResultData(dest)
	return res, fs.OK
}

var _ fs.NodeGetattrer = (*fileNode)(nil)
var _ fs.NodeOpener = (*fileNode)(nil)
var _ fs.NodeReader = (*fileNode)(nil)
