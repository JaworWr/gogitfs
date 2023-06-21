package gitfs

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/error_handler"
	"io"
	"syscall"
)

type allCommitsNode struct {
	repoNode
}

type commitDirStream struct {
	next *object.Commit
	err  error
	iter object.CommitIter
}

func newCommitDirStream(iter object.CommitIter) *commitDirStream {
	ds := &commitDirStream{iter: iter}
	ds.update()
	return ds
}

func (s *commitDirStream) update() {
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
	return s.next != nil
}

func (s *commitDirStream) Next() (entry fuse.DirEntry, errno syscall.Errno) {
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
	entry.Ino = commitNodeMgr.InoStore.GetOrInsert(s.next.Hash, false).Ino
	entry.Mode = fuse.S_IFDIR
	s.update()
	return
}

func (s *commitDirStream) Close() {
	s.iter.Close()
}

func (n *allCommitsNode) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	iter, err := n.repo.CommitObjects()
	if err != nil {
		error_handler.Logging.HandleError(err)
		return nil, syscall.EIO
	}
	return newCommitDirStream(iter), 0
}

func (n *allCommitsNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	hash := plumbing.NewHash(name)
	commit, err := n.repo.CommitObject(hash)
	if err != nil {
		if err == plumbing.ErrObjectNotFound {
			return nil, syscall.ENOENT
		} else {
			error_handler.Logging.HandleError(err)
			return nil, syscall.EIO
		}
	}
	node := newCommitNode(ctx, commit, n)

	out.Mode = syscall.S_IFDIR | 0555
	out.AttrValid = 2 << 62
	out.EntryValid = 2 << 62
	return node, 0
}

//func (h *allCommitsNode) OnAdd(ctx context.Context) {
//	_ = h.commitIter.ForEach(func(commit *object.Commit) error {
//		node := newCommitNode(ctx, commit, h)
//		succ := h.AddChild(commit.Hash.String(), node, false)
//		if !succ {
//			log.Printf("File already exists: %v", commit.Hash.String())
//		}
//		return nil
//	})
//
//	headLink := &fs.MemSymlink{Data: []byte(h.head.Hash().String())}
//	headNode := h.NewPersistentInode(ctx, headLink, fs.StableAttr{Mode: fuse.S_IFLNK})
//	h.AddChild("HEAD", headNode, false)
//}

func newHardlinkCommitListNode(repo *git.Repository) *allCommitsNode {
	node := &allCommitsNode{}
	node.repo = repo
	return node
}

// var _ fs.NodeOnAdder = (*allCommitsNode)(nil)
var _ fs.NodeLookuper = (*allCommitsNode)(nil)
var _ fs.NodeReaddirer = (*allCommitsNode)(nil)
