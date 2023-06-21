package gitfs

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"strings"
)

type commitLogNode struct {
	repoNode
	iter     object.CommitIter
	basePath *string
}

func newCommitLogNode(repo *git.Repository, from plumbing.Hash, linkLevels int) (*commitLogNode, error) {
	opts := &git.LogOptions{From: from}
	log, err := repo.Log(opts)
	if err != nil {
		return nil, err
	}
	node := &commitLogNode{}
	node.repo = repo
	node.iter = log
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
