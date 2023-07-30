package gitfs

import (
	"context"
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
	branch *plumbing.Reference,
	parent repoNodeEmbedder,
) (*object.Commit, *fs.Inode, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if !branch.Name().IsBranch() {
		panic("Reference does not point to a branch!")
	}
	branchName := branch.Name().Short()

	lastHash := m.lastCommitHash[branchName]
	logging.DebugLog.Printf(
		"Branch %v - last: %v, current: %v",
		branchName,
		lastHash.String(),
		branch.Hash().String(),
	)
	overwrite := lastHash != branch.Hash()

	lastCommit, err := parent.embeddedRepoNode().repo.CommitObject(branch.Hash())
	if err != nil {
		return nil, nil, err
	}

	builder := func() (fs.InodeEmbedder, error) {
		logging.InfoLog.Printf(
			"Creating new node for branch %v",
			branchName,
		)
		nodeOpts := commitLogNodeOpts{linkLevels: 0, includeHead: true, symlinkHead: true}
		logNode, err := newCommitLogNode(parent.embeddedRepoNode().repo, lastCommit, nodeOpts)
		if err != nil {
			return nil, err
		}
		m.lastCommitHash[branchName] = lastCommit.Hash
		return logNode, nil
	}
	node, err := m.InodeManager.GetOrInsert(ctx, branchName, fuse.S_IFDIR, parent, builder, overwrite)
	if err != nil {
		return nil, nil, err
	}
	return lastCommit, node, nil
}
