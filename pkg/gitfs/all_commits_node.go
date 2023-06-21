package gitfs

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"log"
)

type allCommitsNode struct {
	repoNode
	head       *plumbing.Reference
	commitIter object.CommitIter
}

func (h *allCommitsNode) OnAdd(ctx context.Context) {
	_ = h.commitIter.ForEach(func(commit *object.Commit) error {
		node := newCommitNode(ctx, commit, h)
		succ := h.AddChild(commit.Hash.String(), node, false)
		if !succ {
			log.Printf("File already exists: %v", commit.Hash.String())
		}
		return nil
	})

	headLink := &fs.MemSymlink{Data: []byte(h.head.Hash().String())}
	headNode := h.NewPersistentInode(ctx, headLink, fs.StableAttr{Mode: fuse.S_IFLNK})
	h.AddChild("HEAD", headNode, false)
}

func newHardlinkCommitListNode(repo *git.Repository) (node *allCommitsNode, err error) {
	head, err := repo.Head()
	if err != nil {
		return
	}
	iter, err := repo.CommitObjects()
	if err != nil {
		return
	}
	node = &allCommitsNode{}
	node.repo = repo
	node.head = head
	node.commitIter = iter
	return
}

var _ fs.NodeOnAdder = (*allCommitsNode)(nil)
