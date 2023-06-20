package gitfs

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type hardlinkCommitListNode struct {
	repoNode
	commits []*object.Commit
}

func newHardlinkCommitListNode(hash *plumbing.Hash, parent repoNodeEmbedder) (node *hardlinkCommitListNode, err error) {
	opts := &git.LogOptions{}
	if hash == nil {
		opts.From = *hash
	} else {
		opts.All = true
	}
	iter, err := parent.embeddedRepoNode().repo.Log(opts)
	if err != nil {
		return
	}
	commits := make([]*object.Commit, 0)
	_ = iter.ForEach(func(commit *object.Commit) error {
		commits = append(commits, commit)
		return nil
	})
	node = &hardlinkCommitListNode{}
	node.repo = parent.embeddedRepoNode().repo
	node.commits = commits
	return
}
