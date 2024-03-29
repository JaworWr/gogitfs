package gitfs

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/inode_manager"
	"gogitfs/pkg/logging"
	"sync"
)

// branchNodeCache extends InodeCache with the information about the last commit on a branch,
// thus making sure that when a new commit is added to the branch, the corresponding node will be updated as well.
type branchNodeCache struct {
	inode_manager.InodeCache
	lock           *sync.Mutex
	lastCommitHash map[string]plumbing.Hash
}

// Init performs initialization. initialIno is passed to InodeCache.Init.
func (m *branchNodeCache) init(initialIno uint64) {
	m.InodeCache.Init(initialIno)
	m.lock = &sync.Mutex{}
	m.lastCommitHash = make(map[string]plumbing.Hash)
}

// getOrInsert returns the Inode corresponding to the given key. If the key is absent or the branch head was changed,
// a new node will be created as in InodeCache. As long as the last commit of the branch remains unchanged,
// subsequent calls will not create a new node.
func (m *branchNodeCache) getOrInsert(
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
		return nil, nil, fmt.Errorf("cannot get last commit of branch %v: %w", branchName, err)
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
	node, err := m.InodeCache.GetOrInsert(ctx, branchName, fuse.S_IFDIR, parent, builder, overwrite)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create node for branch %v: %w", branchName, err)
	}
	return lastCommit, node, nil
}
