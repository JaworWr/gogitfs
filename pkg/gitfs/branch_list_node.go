package gitfs

import (
	"context"
	"errors"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/error_handler"
	"gogitfs/pkg/gitfs/internal/utils"
	"gogitfs/pkg/logging"
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
	next *fuse.DirEntry
	rest <-chan *fuse.DirEntry
	stop chan<- int
}

func readBranchIter(iter storer.ReferenceIter, next chan<- *fuse.DirEntry, stop <-chan int) {
	funcName := logging.CurrentFuncName(0, logging.Package)
	_ = iter.ForEach(func(reference *plumbing.Reference) error {
		logging.DebugLog.Printf(
			"%s: read branch %v",
			funcName,
			reference.Name(),
		)

		var entry fuse.DirEntry
		entry.Name = reference.Name().Short()
		entry.Ino = branchNodeMgr.InoStore.GetOrInsert(reference.Name().Short(), false).Ino
		entry.Mode = fuse.S_IFDIR
		select {
		case <-stop:
			return nil
		case next <- &entry:
		}
		return nil
	})
	close(next)
}

func newBranchDirStream(iter storer.ReferenceIter) *branchDirStream {
	rest := make(chan *fuse.DirEntry, 5)
	stop := make(chan int)
	go readBranchIter(iter, rest, stop)
	stream := &branchDirStream{rest: rest, stop: stop}
	return stream
}

func (s *branchDirStream) HasNext() bool {
	if s.next == nil {
		s.next = <-s.rest
	}
	return s.next != nil
}

func (s *branchDirStream) Next() (entry fuse.DirEntry, errno syscall.Errno) {
	if s.next == nil {
		errno = syscall.ENOENT
		return
	}
	entry = *s.next
	s.next = nil
	return
}

func (s *branchDirStream) Close() {
	s.next = nil
	s.stop <- 1
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
	refName := plumbing.NewBranchReferenceName(name)
	branch, err := n.repo.Reference(refName, false)
	if err != nil {
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
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
