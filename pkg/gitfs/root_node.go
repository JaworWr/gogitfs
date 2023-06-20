package gitfs

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/error_handler"
	"syscall"
)

type RootNode struct {
	repoNode
}

func (n *RootNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	out.Mode = 0555
	out.AttrValid = 10
	out.EntryValid = 10

	switch name {
	case "commits":
		head, err := n.repo.Head()
		if err != nil {
			error_handler.Logging.HandleError(err)
			return nil, syscall.EIO
		}
		node, err := newHardlinkCommitListNode(head, n)
		if err != nil {
			error_handler.Logging.HandleError(err)
			return nil, syscall.EIO
		}
		child := n.NewInode(ctx, node, fs.StableAttr{Ino: rootIno, Mode: fuse.S_IFDIR})
		return child, 0
	default:
		return nil, syscall.ENOENT
	}
}

func (n *RootNode) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	commitEntry := fuse.DirEntry{Mode: 0555, Name: "commits", Ino: rootIno}
	stream := fs.NewListDirStream([]fuse.DirEntry{commitEntry})
	return stream, 0
}

func NewRootNode(path string) (node *RootNode, err error) {
	Init()
	repo, err := git.PlainOpen(path)
	if err != nil {
		return
	}

	node = &RootNode{}
	node.repo = repo
	return
}

var _ fs.NodeLookuper = (*RootNode)(nil)
var _ fs.NodeReaddirer = (*RootNode)(nil)
