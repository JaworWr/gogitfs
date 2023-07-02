package gitfs

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/error_handler"
	"syscall"
)

const BranchValid = 30

type branchListNode struct {
	repoNode
}

func (n *branchListNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	branch, err := n.repo.Branch(name)
	if err != nil {
		if err == git.ErrBranchNotFound {
			return nil, syscall.ENOENT
		} else {
			error_handler.Logging.HandleError(err)
			return nil, syscall.EIO
		}
	}
	commit, node, err := branchNodeMgr.getOrInsert(ctx, branch, n)
	if err != nil {
		error_handler.Logging.HandleError(err)
		return nil, syscall.EIO
	}
	out.AttrValid = BranchValid
	out.EntryValid = BranchValid
	out.Attr = commitAttr(commit)
	out.Mode = fuse.S_IFDIR | 0555
	return node, fs.OK
}

func newBranchListNode(repo *git.Repository) *branchListNode {
	node := &branchListNode{}
	node.repo = repo
	return node
}

var _ fs.NodeLookuper = (*branchListNode)(nil)

//var _ fs.NodeReaddirer = (*branchListNode)(nil)
