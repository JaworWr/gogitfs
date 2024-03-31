package gitfs

import (
	"bytes"
	"context"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/error_handler"
	"gogitfs/pkg/gitfs/internal/utils"
	"gogitfs/pkg/logging"
	"io"
	"syscall"
)

type fileNode struct {
	fs.Inode
	file *object.File
	attr fuse.Attr
	data []byte
}

func newFileNode(file *object.File, attr fuse.Attr) *fileNode {
	node := &fileNode{}
	node.file = file
	node.attr = attr
	node.attr.Mode = 0444
	node.attr.Size = uint64(node.file.Size)
	return node
}

type fileNodeHandle struct{}

func (n *fileNode) GetCallCtx() logging.CallCtx {
	info := utils.NodeCallCtx(n)
	info["name"] = n.file.Name
	info["mode"] = n.file.Mode
	return info
}

func (n *fileNode) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	logging.LogCall(n, nil)
	out.Attr = n.attr
	return fs.OK
}

func (n *fileNode) Open(_ context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	if flags&(syscall.O_RDWR|syscall.O_WRONLY) != 0 {
		return nil, 0, syscall.EROFS
	}

	if n.data == nil {
		reader, err := n.file.Reader()
		if err != nil {
			error_handler.Logging.HandleError(err)
			return nil, 0, syscall.EIO
		}
		defer func(reader io.ReadCloser) {
			_ = reader.Close()
		}(reader)
		buf := &bytes.Buffer{}
		_, err = buf.ReadFrom(reader)
		if err != nil {
			error_handler.Logging.HandleError(err)
			return nil, 0, syscall.EIO
		}
		n.data = buf.Bytes()
	}
	return fileNodeHandle{}, 0, fs.OK
}

var _ fs.NodeGetattrer = (*fileNode)(nil)
var _ fs.NodeOpener = (*fileNode)(nil)
