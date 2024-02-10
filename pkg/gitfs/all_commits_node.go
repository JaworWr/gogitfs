package gitfs

import (
	"context"
	"errors"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/error_handler"
	"gogitfs/pkg/gitfs/internal/utils"
	"gogitfs/pkg/logging"
	"strings"
	"syscall"
	"time"
)

// HeadAttrValid represents expiration time for HEAD symlink attributes
const HeadAttrValid = 30 * time.Second

// allCommitsNode implements a directory containing all commits in the repository.
// Each commit is represented by a directory, whose name is the commit's hash.
// Readdir and Lookup always consider the current state of the repository.
// This way, new commits are included automatically.
type allCommitsNode struct {
	repoNode
	headLink *fs.Inode
}

func (n *allCommitsNode) GetCallCtx() logging.CallCtx {
	return utils.NodeCallCtx(n)
}

// headLinkNode implements the HEAD symlink, i.e. the symlink pointing to the most recent commit on the current branch.
// Readlink always considers the current state of the repository.
type headLinkNode struct {
	repoNode
}

func (n *headLinkNode) GetCallCtx() logging.CallCtx {
	commit, err := headCommit(n)
	if err != nil {
		error_handler.Fatal.HandleError(err)
	}
	info := utils.NodeCallCtx(n)
	info["headHash"] = commit.Hash.String()
	info["headMsg"] = commit.Message
	return info
}

// Readlink returns the path to the current HEAD commit.
func (n *headLinkNode) Readlink(_ context.Context) ([]byte, syscall.Errno) {
	logging.LogCall(n, nil)
	head, err := n.repo.Head()
	if err != nil {
		error_handler.Logging.HandleError(err)
		return nil, syscall.EIO
	}
	return []byte(head.Hash().String()), fs.OK
}

// Getattr returns attributes corresponding to the current HEAD commit.
func (n *headLinkNode) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	logging.LogCall(n, nil)
	attr, err := headAttr(n)
	if err != nil {
		error_handler.Logging.HandleError(err)
		return syscall.EIO
	}
	out.Attr = attr
	out.Attr.Mode = 0555
	out.SetTimeout(HeadAttrValid)
	return fs.OK
}

var _ fs.NodeReadlinker = (*headLinkNode)(nil)
var _ fs.NodeGetattrer = (*headLinkNode)(nil)

// commitDirStream implements an iterator over the contents of the directory representing all commits,
// i.e. the HEAD symlink and the directories representing commits.
// Uses a separate goroutine to actually read the directory.
type commitDirStream struct {
	// HEAD symlink
	headLink *fs.Inode
	// next entry to return
	next *fuse.DirEntry
	// channel returning remaining entries
	rest <-chan *fuse.DirEntry
	// channel indicating that the iteration should stop
	stop chan<- int
}

// readCommitIter reads the commits from `iter`, generates corresponding entries and places them in the channel `next`.
// If a value is read from `stop`, the function returns immediately.
func readCommitIter(iter object.CommitIter, next chan<- *fuse.DirEntry, stop <-chan int) {
	funcName := logging.CurrentFuncName(0, logging.Package)
	err := iter.ForEach(func(commit *object.Commit) error {
		logging.DebugLog.Printf(
			"%s: read commit %v (%v)",
			funcName,
			commit.Hash,
			strings.Replace(commit.Message, "\n", ";", -1),
		)

		var entry fuse.DirEntry
		entry.Name = commit.Hash.String()
		entry.Ino = commitCache.AttrStore.GetOrInsert(commit.Hash.String(), false).Ino
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

// newCommitDirStream creates a new commitDirStream from the commit iterator and an optional HEAD symlink node.
func newCommitDirStream(iter object.CommitIter, headLink *fs.Inode) *commitDirStream {
	rest := make(chan *fuse.DirEntry, 5)
	stop := make(chan int, 1)
	go readCommitIter(iter, rest, stop)
	ds := &commitDirStream{headLink: headLink, rest: rest, stop: stop}
	return ds
}

// HasNext returns true if there are more entries.
func (s *commitDirStream) HasNext() bool {
	logging.LogCall(logging.NilCtx{}, logging.CallCtx{})
	if s.next == nil {
		s.next = <-s.rest
	}
	return s.next != nil || s.headLink != nil
}

// Next returns the next entry. Note that this function depends on HasNext being called first.
func (s *commitDirStream) Next() (entry fuse.DirEntry, errno syscall.Errno) {
	logging.LogCall(logging.NilCtx{}, logging.CallCtx{})
	if s.headLink != nil {
		entry.Name = "HEAD"
		entry.Mode = fuse.S_IFLNK
		entry.Ino = s.headLink.StableAttr().Ino
		s.headLink = nil
		return
	}
	if s.next == nil {
		errno = syscall.ENOENT
		return
	}
	entry = *s.next
	s.next = nil
	return
}

// Close closes the stream and cleans up any resources.
func (s *commitDirStream) Close() {
	logging.LogCall(logging.NilCtx{}, logging.CallCtx{})
	s.next = nil
	s.headLink = nil
	s.stop <- 1
}

// Readdir returns the contents of the directory representing all commits,
// i.e. the HEAD symlink and the directories representing commits.
// The result is based on the current state of the repository.
func (n *allCommitsNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	logging.LogCall(n, nil)
	iter, err := n.repo.CommitObjects()
	if err != nil {
		error_handler.Logging.HandleError(err)
		return nil, syscall.EIO
	}
	return newCommitDirStream(iter, n.getHeadLinkNode(ctx)), fs.OK
}

// Lookup returns a node representing the commit with the given hash, or the HEAD symlink if `name == "HEAD"`.
// The result is based on the current state of the repository.
func (n *allCommitsNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	logging.LogCall(n, logging.CallCtx{"name": name})
	var err error
	if name == "HEAD" {
		headLink := n.getHeadLinkNode(ctx)
		out.Attr, err = headAttr(n)
		if err != nil {
			error_handler.Logging.HandleError(err)
			return nil, syscall.EIO
		}
		out.Mode = fuse.S_IFLNK | 0555
		out.SetAttrTimeout(HeadAttrValid)
		return headLink, fs.OK
	}

	hash := plumbing.NewHash(name)
	commit, err := n.repo.CommitObject(hash)
	if err != nil {
		if errors.Is(err, plumbing.ErrObjectNotFound) {
			logging.WarningLog.Printf("Commit %v not found", name)
			return nil, syscall.ENOENT
		} else {
			error_handler.Logging.HandleError(err)
			return nil, syscall.EIO
		}
	}
	node := newCommitNode(ctx, commit, n)

	out.Attr = utils.CommitAttr(commit)
	out.Mode = syscall.S_IFDIR | 0555
	return node, fs.OK
}

// Getattr returns attributes corresponding to the current HEAD commit.
func (n *allCommitsNode) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	logging.LogCall(n, nil)
	attr, err := headAttr(n)
	if err != nil {
		error_handler.Logging.HandleError(err)
		return syscall.EIO
	}
	out.Attr = attr
	out.Attr.Mode = 0555
	out.SetTimeout(HeadAttrValid)
	return fs.OK
}

func newAllCommitsNode(repo *git.Repository) *allCommitsNode {
	node := &allCommitsNode{}
	node.repo = repo
	return node
}

func (n *allCommitsNode) getHeadLinkNode(ctx context.Context) *fs.Inode {
	if n.headLink == nil {
		headLink := &headLinkNode{}
		headLink.repo = n.repo
		hlNode := n.NewInode(ctx, headLink, fs.StableAttr{Mode: fuse.S_IFLNK})
		n.headLink = hlNode
	}
	return n.headLink
}

var _ fs.NodeLookuper = (*allCommitsNode)(nil)
var _ fs.NodeReaddirer = (*allCommitsNode)(nil)
var _ fs.NodeGetattrer = (*allCommitsNode)(nil)
