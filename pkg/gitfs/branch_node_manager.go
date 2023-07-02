package gitfs

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/inode_manager"
)

type branchNodeManager struct {
	inode_manager.InodeManager
	lastCommit map[string]plumbing.Hash
}

func newBranchNodeManager(initialIno uint64) *branchNodeManager {
	nodeMgr := &branchNodeManager{}
	nodeMgr.InodeManager.Init(initialIno)
	nodeMgr.lastCommit = make(map[string]plumbing.Hash)
	return nodeMgr
}

func (n *branchNodeManager) getOrInsert(
	ctx context.Context,
	branch *config.Branch,
	parent repoNodeEmbedder,
) (*object.Commit, *fs.Inode, error) {
	lastHash := n.lastCommit[branch.Name]
	commit, err := getBranchCommit(parent.embeddedRepoNode().repo, branch)
	if err != nil {
		return nil, nil, err
	}
	overwrite := lastHash != commit.Hash
	nodeOpts := commitLogNodeOpts{linkLevels: 0, includeHead: true, symlinkHead: true}
	logNode, err := newCommitLogNode(parent.embeddedRepoNode().repo, commit, nodeOpts)
	if err != nil {
		return nil, nil, err
	}
	builder := func() fs.InodeEmbedder {
		return logNode
	}
	node := n.InodeManager.GetOrInsert(ctx, branch.Name, fuse.S_IFDIR, parent, builder, overwrite)
	return commit, node, nil
}

func getBranchCommit(repo *git.Repository, branch *config.Branch) (*object.Commit, error) {
	ref, err := repo.Reference(branch.Merge, true)
	if err != nil {
		return nil, err
	}
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}
	return commit, nil
}
