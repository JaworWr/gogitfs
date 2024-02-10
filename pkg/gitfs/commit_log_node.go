package gitfs

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/gitfs/internal/utils"
	"gogitfs/pkg/logging"
	"path"
	"syscall"
)

// commitLogNode represents a commit log, or any other subset of repo commits.
// Each commit is represented as a symlink or a hardlink, whose name and attributes correspond to those of
// the actual directory representing the commit (in particular, the symlink's name is the hash of the commit).
type commitLogNode struct {
	repoNode
	// from is the start commit of the log
	from *object.Commit
	// iter is the iterator over commits, which may or may not contain from.
	iter object.CommitIter
	// basePath is the path of the actual commit directory. All symlinks will link to its subdirectories.
	// If set to nil, the links will be hardlinks instead.
	basePath *string
	// attr represents attributes of the inode corresponding to this object.
	attr fuse.Attr
	// If true, the head commit (represented by from) will be included in the represented directory.
	// Otherwise, it will be skipped.
	includeHead bool
	// If true, a symlink to the head commit called "HEAD" will be created - requires includeHead.
	symlinkHead bool
}

func (n *commitLogNode) GetCallCtx() logging.CallCtx {
	info := utils.NodeCallCtx(n)
	info["from"] = n.from.Hash.String()
	if n.basePath == nil {
		info["basepath"] = "<nil>"
	} else {
		info["basepath"] = *n.basePath
	}
	info["includeHead"] = n.includeHead
	info["symlinkHead"] = n.symlinkHead
	return info
}

// Getattr returns attributes corresponding to the head commit of the log.
func (n *commitLogNode) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	logging.LogCall(n, nil)
	out.Attr = n.attr
	return fs.OK
}

// commitSymlink creates a symlink node for a given commit object
func commitSymlink(commit *object.Commit, basePath *string) *fs.MemSymlink {
	attr := utils.CommitAttr(commit)
	attr.Mode = 0555
	p := commit.Hash.String()
	if basePath != nil {
		p = path.Join(*basePath, p)
	}
	link := &fs.MemSymlink{Attr: attr, Data: []byte(p)}
	return link
}

// OnAdd populates the inode with links corresponding to commits and the optional HEAD symlink.
func (n *commitLogNode) OnAdd(ctx context.Context) {
	logging.LogCall(n, nil)
	if n.basePath == nil {
		n.addHardlinks(ctx)
	} else {
		n.addSymlinks(ctx, *n.basePath)
	}
	if n.symlinkHead {
		link := commitSymlink(n.from, nil)
		node := n.NewPersistentInode(ctx, link, fs.StableAttr{Mode: fuse.S_IFLNK})
		n.AddChild("HEAD", node, false)
	}
}

func (n *commitLogNode) addHardlinks(ctx context.Context) {
	_ = n.iter.ForEach(func(commit *object.Commit) error {
		if !n.includeHead && commit.Hash == n.from.Hash {
			return nil
		}
		node := newCommitNode(ctx, commit, n)
		succ := n.AddChild(commit.Hash.String(), node, false)
		if !succ {
			logging.WarningLog.Printf("Duplicate commit node: %v\n", commit.Hash.String())
		}
		return nil
	})
}

func (n *commitLogNode) addSymlinks(ctx context.Context, basePath string) {
	_ = n.iter.ForEach(func(commit *object.Commit) error {
		if !n.includeHead && commit.Hash == n.from.Hash {
			return nil
		}
		link := commitSymlink(commit, &basePath)
		node := n.NewPersistentInode(ctx, link, fs.StableAttr{Mode: fuse.S_IFLNK})
		succ := n.AddChild(commit.Hash.String(), node, false)
		if !succ {
			logging.WarningLog.Printf("Duplicate commit node: %v\n", commit.Hash.String())
		}
		return nil
	})
}

// getBasePath creates a path that leads linkLevels directories up. If linkLevels == 0, returns nil,
// which signalizes that hardlinks should be used.
func getBasePath(linkLevels int) (basePath *string) {
	if linkLevels == 0 {
		basePath = nil
	} else {
		elems := make([]string, linkLevels)
		for i := range elems {
			elems[i] = ".."
		}
		p := path.Join(elems...)
		basePath = &p
	}
	return
}

type commitLogNodeOpts struct {
	// linkLevels determines how many directories upward in the hierarchy the symlinks should point.
	// Setting this value to 0 signalizes that hardlinks should be used.
	linkLevels int
	// If true, the head commit (represented by from) will be included in the represented directory.
	// Otherwise, it will be skipped.
	includeHead bool
	// If true, a symlink to the head commit called "HEAD" will be created - requires includeHead.
	symlinkHead bool
}

// newCommitLogNode creates a node commitLogNode from the git log starting at the commit `from`.
func newCommitLogNode(repo *git.Repository, from *object.Commit, nodeOpts commitLogNodeOpts) (*commitLogNode, error) {
	opts := &git.LogOptions{From: from.Hash}
	iter, err := repo.Log(opts)
	if err != nil {
		return nil, fmt.Errorf("cannot get commit log: %w", err)
	}
	node := newCommitLogNodeFromIter(iter, repo, from, nodeOpts)
	return node, nil
}

// newCommitLogNode creates a node commitLogNode starting from an arbitrary CommitIter,
// which may or may not contain `from`.
func newCommitLogNodeFromIter(
	iter object.CommitIter,
	repo *git.Repository,
	from *object.Commit,
	nodeOpts commitLogNodeOpts,
) *commitLogNode {
	node := &commitLogNode{}
	node.repo = repo
	node.from = from
	node.iter = iter
	node.includeHead = nodeOpts.includeHead
	node.symlinkHead = nodeOpts.symlinkHead
	node.attr = utils.CommitAttr(from)
	node.attr.Mode = 0555
	node.basePath = getBasePath(nodeOpts.linkLevels)
	return node
}

var _ fs.NodeOnAdder = (*commitLogNode)(nil)
var _ fs.NodeGetattrer = (*commitLogNode)(nil)
