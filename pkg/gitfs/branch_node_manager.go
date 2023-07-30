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
	"gogitfs/pkg/logging"
	"sync"
)

type branchNodeManager struct {
	inode_manager.InodeManager
	lock           *sync.Mutex
	lastCommitHash map[string]plumbing.Hash
}

func (m *branchNodeManager) init(initialIno uint64) {
	m.InodeManager.Init(initialIno)
	m.lock = &sync.Mutex{}
	m.lastCommitHash = make(map[string]plumbing.Hash)
}

func (m *branchNodeManager) getOrInsert(
	ctx context.Context,
	branch *config.Branch,
	parent repoNodeEmbedder,
) (*object.Commit, *fs.Inode, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	lastHash := m.lastCommitHash[branch.Name]
	lastCommit, err := getBranchCommit(parent.embeddedRepoNode().repo, branch)
	if err != nil {
		return nil, nil, err
	}
	logging.DebugLog.Printf(
		"Branch %v - last: %v, current: %v",
		branch.Name,
		lastHash.String(),
		lastCommit.Hash.String(),
	)
	overwrite := lastHash != lastCommit.Hash
	if err != nil {
		return nil, nil, err
	}
	builder := func() (fs.InodeEmbedder, error) {
		logging.InfoLog.Printf(
			"Creating new node for branch %v",
			branch.Name,
		)
		nodeOpts := commitLogNodeOpts{linkLevels: 0, includeHead: true, symlinkHead: true}
		logNode, err := newCommitLogNode(parent.embeddedRepoNode().repo, lastCommit, nodeOpts)
		if err != nil {
			return nil, err
		}
		m.lastCommitHash[branch.Name] = lastCommit.Hash
		return logNode, nil
	}
	node, err := m.InodeManager.GetOrInsert(ctx, branch.Name, fuse.S_IFDIR, parent, builder, overwrite)
	if err != nil {
		return nil, nil, err
	}
	return lastCommit, node, nil
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
