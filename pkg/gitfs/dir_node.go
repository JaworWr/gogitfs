package gitfs

import (
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/gitfs/internal/utils"
	"gogitfs/pkg/logging"
)

type dirNode struct {
	repoNode
	entry *object.TreeEntry
	tree  *object.Tree
	attr  fuse.Attr
}

func (n *dirNode) GetCallCtx() logging.CallCtx {
	info := utils.NodeCallCtx(n)
	info["mode"] = n.entry.Mode
	return info
}
