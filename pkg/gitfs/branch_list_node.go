package gitfs

import (
	"github.com/go-git/go-git/v5"
)

type branchListNode struct {
	repoNode
}

func newBranchListNode(repo *git.Repository) *branchListNode {
	node := &branchListNode{}
	node.repo = repo
	return node
}

//var _ fs.NodeLookuper = (*branchListNode)(nil)
//var _ fs.NodeReaddirer = (*branchListNode)(nil)
