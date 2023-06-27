package gitfs

import (
	"context"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/inode_manager"
)

type branchNodeManager struct {
	inodeMgr   *inode_manager.InodeManager
	lastCommit map[string]plumbing.Hash
}

func newBranchNodeManager(initialIno uint64) *branchNodeManager {
	nodeMgr := &branchNodeManager{}
	nodeMgr.inodeMgr = inode_manager.NewInodeManager(initialIno)
	nodeMgr.lastCommit = make(map[string]plumbing.Hash)
	return nodeMgr
}

func (n *branchNodeManager) getOrInsert(
	ctx context.Context,
	branch string,
	commit *object.Commit,
	parent repoNodeEmbedder,
) (*fs.Inode, error) {
	lastHash := n.lastCommit[branch]
	overwrite := lastHash != commit.Hash
	nodeOpts := commitLogNodeOpts{linkLevels: 0}
	logNode, err := newCommitLogNode(parent.embeddedRepoNode().repo, commit, nodeOpts)
	if err != nil {
		return nil, err
	}
	builder := func() fs.InodeEmbedder {
		return logNode
	}
	node := n.inodeMgr.GetOrInsert(ctx, branch, fuse.S_IFDIR, parent, builder, overwrite)
	return node, nil
}
