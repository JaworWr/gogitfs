package gitfs

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/error_handler"
)

type RootNode struct {
	repoNode
}

func (n *RootNode) OnAdd(ctx context.Context) {
	node, err := newHardlinkCommitListNode(n.repo)
	if err != nil {
		error_handler.Fatal.HandleError(err)
		return
	}
	child := n.NewPersistentInode(ctx, node, fs.StableAttr{Mode: fuse.S_IFDIR})
	n.AddChild("commits", child, false)
}

func NewRootNode(path string) (node *RootNode, err error) {
	Init()
	repo, err := git.PlainOpen(path)
	if err != nil {
		return
	}

	node = &RootNode{}
	node.repo = repo
	return
}

var _ fs.NodeOnAdder = (*RootNode)(nil)
