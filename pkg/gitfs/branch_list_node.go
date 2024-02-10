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

// BranchValid represents expiration time for branch nodes
const BranchValid = 30 * time.Second

// branchListNode represents the list of all branches. Each branch is represented as a directory named after the branch.
// Readdir and Lookup always consider the current state of the repository.
type branchListNode struct {
	repoNode
}

func (n *branchListNode) GetCallCtx() logging.CallCtx {
	return utils.NodeCallCtx(n)
}

// branchDirStream implements an iterator over the contents of the branch directory.
// Uses a separate goroutine to actually read the directory.
type branchDirStream struct {
	next *fuse.DirEntry
	rest <-chan *fuse.DirEntry
	stop chan<- int
}

// readBranchIter reads the branch references from `iter`, generates corresponding entries
// and places them in the channel `next`. If a value is read from `stop`, the function returns immediately.
func readBranchIter(iter storer.ReferenceIter, next chan<- *fuse.DirEntry, stop <-chan int) {
	funcName := logging.CurrentFuncName(0, logging.Package)
	err := iter.ForEach(func(reference *plumbing.Reference) error {
		logging.DebugLog.Printf(
			"%s: read branch %v",
			funcName,
			reference.Name(),
		)

		var entry fuse.DirEntry
		entry.Name = reference.Name().Short()
		entry.Ino = branchCache.AttrStore.GetOrInsert(reference.Name().Short(), false).Ino
		entry.Mode = fuse.S_IFDIR
		select {
		case <-stop:
			return nil
		case next <- &entry:
		}
		return nil
	})
	if err != nil {
		error_handler.Logging.HandleError(err)
	}
	close(next)
}

// newBranchDirStream creates a new branchDirStream from the branch reference iterator.
func newBranchDirStream(iter storer.ReferenceIter) *branchDirStream {
	rest := make(chan *fuse.DirEntry, 5)
	stop := make(chan int, 1)
	go readBranchIter(iter, rest, stop)
	stream := &branchDirStream{rest: rest, stop: stop}
	return stream
}

// HasNext returns true if there are more entries.
func (s *branchDirStream) HasNext() bool {
	if s.next == nil {
		s.next = <-s.rest
	}
	return s.next != nil
}

// Next returns the next entry. Note that this function depends on HasNext being called first.
func (s *branchDirStream) Next() (entry fuse.DirEntry, errno syscall.Errno) {
	if s.next == nil {
		errno = syscall.ENOENT
		return
	}
	entry = *s.next
	s.next = nil
	return
}

// Close closes the stream and cleans up any resources.
func (s *branchDirStream) Close() {
	s.next = nil
	s.stop <- 1
}

// Readdir returns the contents of the directory representing all branches.
// The result is based on the current state of the repository.
func (n *branchListNode) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	logging.LogCall(n, nil)
	iter, err := n.repo.Branches()
	if err != nil {
		error_handler.Logging.HandleError(err)
		return nil, syscall.EIO
	}
	return newBranchDirStream(iter), fs.OK
}

// Lookup returns a node representing the branch with the given name.
// The result is based on the current state of the repository.
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
	commit, node, err := branchCache.getOrInsert(ctx, branch, n)
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

// Getattr returns attributes corresponding to the current HEAD commit.
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
