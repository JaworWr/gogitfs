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
	"io"
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
	next     *object.Commit
	err      error
	iter     object.CommitIter
}

func newCommitDirStream(iter object.CommitIter, headLink *fs.Inode) *commitDirStream {
	ds := &commitDirStream{iter: iter, headLink: headLink}
	return ds
}

func (s *commitDirStream) update() {
	s.headLink = nil
	next, err := s.iter.Next()
	if err != nil {
		next = nil
		if err != io.EOF {
			s.err = err
		}
	}
	s.next = next
}

func (s *commitDirStream) HasNext() bool {
	return s.next != nil || s.headLink != nil
}

func (s *commitDirStream) Next() (entry fuse.DirEntry, errno syscall.Errno) {
	if s.headLink != nil {
		entry.Name = "HEAD"
		entry.Mode = fuse.S_IFLNK
		entry.Ino = s.headLink.StableAttr().Ino
		s.update()
		return
	}
	if s.err != nil {
		error_handler.Logging.HandleError(s.err)
		errno = syscall.EIO
		return
	}
	if s.next == nil {
		errno = syscall.ENOENT
		return
	}

	entry.Name = s.next.Hash.String()
	entry.Ino = commitNodeMgr.InoStore.GetOrInsert(s.next.Hash.String(), false).Ino
	entry.Mode = fuse.S_IFDIR
	s.update()
	return
}

func (s *commitDirStream) Close() {
	s.next = nil
	s.headLink = nil
	s.iter.Close()
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
