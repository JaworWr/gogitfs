package gitfs

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/error_handler"
	"gogitfs/pkg/gitfs/internal/utils"
	"gogitfs/pkg/logging"
	"io"
	"syscall"
	"time"
)

const BranchValid = 30 * time.Second

type branchListNode struct {
	repoNode
}

func (n *branchListNode) GetCallCtx() logging.CallCtx {
	return utils.NodeCallCtx(n)
}

type branchDirStream struct {
	next *plumbing.Reference
	err  error
	iter storer.ReferenceIter
}

func newBranchDirStream(iter storer.ReferenceIter) *branchDirStream {
	stream := &branchDirStream{iter: iter}
	stream.update()
	return stream
}

func (s *branchDirStream) update() {
	next, err := s.iter.Next()
	if err != nil {
		next = nil
		if err != io.EOF {
			s.err = err
		}
	}
	s.next = next
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
	entry.Ino = branchNodeMgr.InoStore.GetOrInsert(s.next.Hash().String(), false).Ino
	entry.Mode = fuse.S_IFDIR
	s.update()
	return
}

func (s *branchDirStream) Close() {
	s.next = nil
	s.iter.Close()
}

func (n *branchListNode) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	logging.LogCall(n, nil)
	iter, err := n.repo.Branches()
	if err != nil {
		error_handler.Logging.HandleError(err)
		return nil, syscall.EIO
	}
	return newBranchDirStream(iter), fs.OK
}

func (n *branchListNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	logging.LogCall(n, logging.CallCtx{"name": name})
	branch, err := n.repo.Branch(name)
	if err != nil {
		if err == git.ErrBranchNotFound {
			logging.WarningLog.Printf("Branch %v not found", name)
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
	out.SetAttrTimeout(BranchValid)
	out.SetEntryTimeout(BranchValid)
	out.Attr = utils.CommitAttr(commit)
	out.Mode = fuse.S_IFDIR | 0555
	return node, fs.OK
}

func (n *branchListNode) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	logging.LogCall(n, nil)
	attr, err := headAttr(n)
	if err != nil {
		error_handler.Logging.HandleError(err)
		return syscall.EIO
	}
	out.Attr = attr
	out.Attr.Mode = 0555
	return fs.OK
}

func newBranchListNode(repo *git.Repository) *branchListNode {
	node := &branchListNode{}
	node.repo = repo
	return node
}

var _ fs.NodeLookuper = (*branchListNode)(nil)
var _ fs.NodeReaddirer = (*branchListNode)(nil)
var _ fs.NodeGetattrer = (*branchListNode)(nil)
