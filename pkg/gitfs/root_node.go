package gitfs

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	"syscall"
)

type RootNode struct {
	fs.Inode
}

func (n *RootNode) OnAdd(ctx context.Context) {
	child := fs.MemRegularFile{}
	child.Data = []byte("Hello there!")
	childNode := n.NewPersistentInode(ctx, &child, fs.StableAttr{Mode: syscall.S_IFREG})
	n.AddChild("hello.txt", childNode, false)
}

var _ fs.NodeOnAdder = (*RootNode)(nil)
