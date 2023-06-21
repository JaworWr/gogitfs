package gitfs

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/error_handler"
	"syscall"
)

type allCommitsNode struct {
	repoNode
}

func (h *allCommitsNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	//TODO implement me
	panic("implement me")
}

func (h *allCommitsNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	hash := plumbing.NewHash(name)
	commit, err := h.repo.CommitObject(hash)
	if err != nil {
		if err == plumbing.ErrObjectNotFound {
			return nil, syscall.ENOENT
		} else {
			error_handler.Logging.HandleError(err)
			return nil, syscall.EIO
		}
	}
	node := newCommitNode(ctx, commit, h)

	out.Mode = syscall.S_IFDIR | 0555
	out.AttrValid = 2 << 63
	out.EntryValid = 2 << 63
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
