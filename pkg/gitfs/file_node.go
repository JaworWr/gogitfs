package gitfs

import (
	"github.com/go-git/go-git/v5/plumbing/object"
	"gogitfs/pkg/gitfs/internal/utils"
	"gogitfs/pkg/logging"
)

type FileNode struct {
	repoNode
	file *object.File
	data []byte
}

type fileNodeHandle struct {
	data []byte
}

func (n *FileNode) GetCallCtx() logging.CallCtx {
	info := utils.NodeCallCtx(n)
	info["name"] = n.file.Name
	info["mode"] = n.file.Mode
	return info
}
