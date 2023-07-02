package gitfs

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/error_handler"
	"syscall"
)

const BranchValid = 30

type branchListNode struct {
	repoNode
}

type branchDirStream struct {
	next *plumbing.Reference
	err  error
	iter storer.ReferenceIter
}

func (s *branchDirStream) update() {

}

func (s *branchDirStream) HasNext() bool {
	return s.next != nil
}

func (s *branchDirStream) Next() (entry fuse.DirEntry, errno syscall.Errno) {
	if s.err != nil {
		error_handler.Logging.HandleError(s.err)
		errno = syscall.EIO
		return
	}
	if s.next == nil {
		errno = syscall.ENOENT
		return
	}
	entry.Name = s.next.Name().Short()
	//entry.Ino
	entry.Mode = fuse.S_IFDIR
	s.update()
	return
}

func (s *branchDirStream) Close() {
	s.next = nil
	s.iter.Close()
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