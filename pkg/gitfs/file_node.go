package gitfs

import (
	"context"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/gitfs/internal/utils"
	"gogitfs/pkg/logging"
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

func (n *fileNode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	logging.LogCall(n, nil)
	out.Attr = n.attr
	return syscall.F_OK
}

var _ fs.NodeGetattrer = (*fileNode)(nil)
