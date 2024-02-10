// Package gitfs contains implementations of the inodes used by the filesystem.
package gitfs

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/gitfs/internal/utils"
	"gogitfs/pkg/logging"
)

type repoNode struct {
	fs.Inode
	repo *git.Repository
}

type repoNodeEmbedder interface {
	fs.InodeEmbedder
	embeddedRepoNode() *repoNode
}

func (n *repoNode) embeddedRepoNode() *repoNode {
	return n
}

func headCommit(n repoNodeEmbedder) (commit *object.Commit, err error) {
	repo := n.embeddedRepoNode().repo
	head, err := repo.Head()
	if err != nil {
		err = fmt.Errorf("cannot get HEAD object: %w", err)
		return
	}
	logging.DebugLog.Printf("HEAD points to %v", head.Hash().String())
	commit, err = repo.CommitObject(head.Hash())
	if err != nil {
		err = fmt.Errorf("cannot get commit object: %w", err)
	}
	return
}

func headAttr(n repoNodeEmbedder) (attr fuse.Attr, err error) {
	commit, err := headCommit(n)
	if err != nil {
		return
	}
	attr = utils.CommitAttr(commit)
	return
}
