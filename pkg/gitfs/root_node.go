package gitfs

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
)

type RootNode struct {
	repoNode
	commit *object.Commit
}

func NewRootNode(path string) (node *RootNode, err error) {
	Init()
	repo, err := git.PlainOpen(path)
	if err != nil {
		return
	}
	ref, err := repo.Head()
	if err != nil {
		return
	}
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return
	}
	node = &RootNode{}
	node.repo = repo
	node.commit = commit
	return
}

func (n *RootNode) OnAdd(ctx context.Context) {
	child := newCommitNode(ctx, n.commit, n)
	n.AddChild("HEAD", child, false)
}

var _ fs.NodeOnAdder = (*RootNode)(nil)
