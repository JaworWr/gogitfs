package gitfs

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"log"
	"strings"
	"syscall"
)

type commitLogNode struct {
	repoNode
	from     plumbing.Hash
	iter     object.CommitIter
	basePath *string
	attr     fuse.Attr
}

func (n *commitLogNode) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Attr = n.attr
	out.AttrValid = 2 << 62
	return fs.OK
}

func (n *commitLogNode) OnAdd(ctx context.Context) {
	if n.basePath == nil {
		n.addHardlinks(ctx)
	} else {
		n.addSymlinks(ctx, *n.basePath)
	}
}

func (n *commitLogNode) addHardlinks(ctx context.Context) {
	_ = n.iter.ForEach(func(commit *object.Commit) error {
		if commit.Hash == n.from {
			return nil
		}
		node := newCommitNode(ctx, commit, n)
		succ := n.AddChild(commit.Hash.String(), node, false)
		if !succ {
			log.Printf("Duplicate commit node: %v\n", commit.Hash.String())
		}
		return nil
	})
}

func (n *commitLogNode) addSymlinks(ctx context.Context, basePath string) {
	_ = n.iter.ForEach(func(commit *object.Commit) error {
		if commit.Hash == n.from {
			return nil
		}
		attr := commitAttr(commit)
		attr.Mode = 0555
		path := fmt.Sprintf("%v/%v", basePath, commit.Hash.String())
		link := &fs.MemSymlink{Attr: attr, Data: []byte(path)}
		node := n.NewPersistentInode(ctx, link, fs.StableAttr{Mode: fuse.S_IFLNK})
		succ := n.AddChild(commit.Hash.String(), node, false)
		if !succ {
			log.Printf("Duplicate commit node: %v\n", commit.Hash.String())
		}
		return nil
	})
}

func newCommitLogNode(repo *git.Repository, from *object.Commit, linkLevels int) (*commitLogNode, error) {
	opts := &git.LogOptions{From: from.Hash}
	iter, err := repo.Log(opts)
	if err != nil {
		return nil, err
	}
	node := &commitLogNode{}
	node.repo = repo
	node.from = from.Hash
	node.iter = iter
	node.attr = commitAttr(from)
	node.attr.Mode = 0555
	if linkLevels == 0 {
		node.basePath = nil
	} else {
		elems := make([]string, linkLevels)
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
