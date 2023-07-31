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

const HeadAttrValid = 30 * time.Second

type allCommitsNode struct {
	repoNode
	headLink *fs.Inode
}

func (n *allCommitsNode) GetCallCtx() logging.CallCtx {
	return utils.NodeCallCtx(n)
}

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

func (n *headLinkNode) Readlink(_ context.Context) ([]byte, syscall.Errno) {
	logging.LogCall(n, nil)
	head, err := n.repo.Head()
	if err != nil {
		error_handler.Logging.HandleError(err)
		return nil, syscall.EIO
	}
	return []byte(head.Hash().String()), fs.OK
}

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

type commitDirStream struct {
	headLink *fs.Inode
	next     *fuse.DirEntry
	rest     <-chan *fuse.DirEntry
	stop     chan<- int
}

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
		entry.Ino = commitNodeMgr.InoStore.GetOrInsert(commit.Hash.String(), false).Ino
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

func newCommitDirStream(iter object.CommitIter, headLink *fs.Inode) *commitDirStream {
	rest := make(chan *fuse.DirEntry, 5)
	stop := make(chan int)
	go readCommitIter(iter, rest, stop)
	ds := &commitDirStream{headLink: headLink, rest: rest, stop: stop}
	return ds
}

func (s *commitDirStream) HasNext() bool {
	if s.next == nil {
		s.next = <-s.rest
	}
	return s.next != nil || s.headLink != nil
}

func (s *commitDirStream) Next() (entry fuse.DirEntry, errno syscall.Errno) {
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

func (s *commitDirStream) Close() {
	s.next = nil
	s.headLink = nil
	s.stop <- 1
}

func (n *allCommitsNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	logging.LogCall(n, nil)
	iter, err := n.repo.CommitObjects()
	if err != nil {
		error_handler.Logging.HandleError(err)
		return nil, syscall.EIO
	}
	return newCommitDirStream(iter, n.getHeadLinkNode(ctx)), fs.OK
}

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
