package gitfs

import (
	"github.com/go-git/go-git/v5"
	"github.com/hanwen/go-fuse/v2/fs"
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
