package gitfs

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/logging"
	"strings"
	"syscall"
)

type commitLogNode struct {
	repoNode
	from        *object.Commit
	iter        object.CommitIter
	basePath    *string
	attr        fuse.Attr
	includeHead bool
	symlinkHead bool
}

func (n *commitLogNode) GetCallCtx() map[string]string {
	info := make(map[string]string)
	info["from"] = n.from.Hash.String()
	if n.basePath == nil {
		info["basepath"] = "<nil>"
	} else {
		info["basepath"] = *n.basePath
	}
	info["includeHead"] = logging.BoolToStr(n.includeHead)
	info["symlinkHead"] = logging.BoolToStr(n.symlinkHead)
	return info
}

func (n *commitLogNode) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	logging.LogCall(n, nil)
	out.Attr = n.attr
	return fs.OK
}

func (n *commitLogNode) OnAdd(ctx context.Context) {
	logging.LogCall(n, nil)
	if n.basePath == nil {
		n.addHardlinks(ctx)
	} else {
		n.addSymlinks(ctx, *n.basePath)
	}
	if n.symlinkHead {
		attr := commitAttr(n.from)
		attr.Mode = 0555
		path := n.from.Hash.String()
		link := &fs.MemSymlink{Attr: attr, Data: []byte(path)}
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
		attr := commitAttr(commit)
		attr.Mode = 0555
		path := fmt.Sprintf("%v/%v", basePath, commit.Hash.String())
		link := &fs.MemSymlink{Attr: attr, Data: []byte(path)}
		node := n.NewPersistentInode(ctx, link, fs.StableAttr{Mode: fuse.S_IFLNK})
		succ := n.AddChild(commit.Hash.String(), node, false)
		if !succ {
			logging.WarningLog.Printf("Duplicate commit node: %v\n", commit.Hash.String())
		}
		return nil
	})
}

type commitLogNodeOpts struct {
	linkLevels  int
	includeHead bool
	symlinkHead bool
}

func newCommitLogNode(repo *git.Repository, from *object.Commit, nodeOpts commitLogNodeOpts) (*commitLogNode, error) {
	opts := &git.LogOptions{From: from.Hash}
	iter, err := repo.Log(opts)
	if err != nil {
		return nil, err
	}
	node := &commitLogNode{}
	node.repo = repo
	node.from = from
	node.iter = iter
	node.includeHead = nodeOpts.includeHead
	node.symlinkHead = nodeOpts.symlinkHead
	node.attr = commitAttr(from)
	node.attr.Mode = 0555
	if nodeOpts.linkLevels == 0 {
		node.basePath = nil
	} else {
		elems := make([]string, nodeOpts.linkLevels)
		for i := range elems {
			elems[i] = ".."
		}
		basePath := strings.Join(elems, "/")
		node.basePath = &basePath
	}
	return node, nil
}

var _ fs.NodeOnAdder = (*commitLogNode)(nil)
var _ fs.NodeGetattrer = (*commitLogNode)(nil)
