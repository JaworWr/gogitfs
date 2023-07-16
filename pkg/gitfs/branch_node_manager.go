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
	"sync"
)

type branchNodeManager struct {
	inode_manager.InodeManager
	lock       *sync.Mutex
	lastCommit map[string]plumbing.Hash
}

func (m *branchNodeManager) init(initialIno uint64) {
	m.InodeManager.Init(initialIno)
	m.lock = &sync.Mutex{}
	m.lastCommit = make(map[string]plumbing.Hash)
}

func (m *branchNodeManager) getOrInsert(
	ctx context.Context,
	branch *config.Branch,
	parent repoNodeEmbedder,
) (*object.Commit, *fs.Inode, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	lastHash := m.lastCommit[branch.Name]
	commit, err := getBranchCommit(parent.embeddedRepoNode().repo, branch)
	if err != nil {
		return nil, nil, err
	}
	overwrite := lastHash != commit.Hash
	if err != nil {
		return nil, nil, err
	}
	builder := func() (fs.InodeEmbedder, error) {
		nodeOpts := commitLogNodeOpts{linkLevels: 0, includeHead: true, symlinkHead: true}
		logNode, err := newCommitLogNode(parent.embeddedRepoNode().repo, commit, nodeOpts)
		if err != nil {
			return nil, err
		}
		return logNode, nil
	}
	node, err := m.InodeManager.GetOrInsert(ctx, branch.Name, fuse.S_IFDIR, parent, builder, overwrite)
	if err != nil {
		return nil, nil, err
	}
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
