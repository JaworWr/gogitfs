package gitfs

import (
	"context"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type commitNode struct {
	repoNode
	commit object.Commit
}

func newCommitNode(ctx context.Context, commit object.Commit, parent repoNodeEmbedder) *fs.Inode {
	builder := func() fs.InodeEmbedder {
		node := commitNode{commit: commit}
		node.repo = parent.embeddedRepoNode().repo
		return &node
	}
	return commitNodeMgr.GetOrInsert(ctx, commit.Hash, fuse.S_IFDIR, parent, builder, false)
}
