package gitfs

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/error_handler"
	"gogitfs/pkg/gitfs/internal/utils"
	"gogitfs/pkg/logging"
	"syscall"
)

type RootNode struct {
	repoNode
}

func (n *RootNode) GetCallCtx() logging.CallCtx {
	return utils.NodeCallCtx(n)
}

func (n *RootNode) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	logging.LogCall(n, nil)
	attr, err := headAttr(n)
	if err != nil {
		error_handler.Logging.HandleError(err)
		return syscall.EIO
	}
	out.Attr = attr
	out.Attr.Mode = 0555
	return fs.OK
}

func (n *RootNode) OnAdd(ctx context.Context) {
	logging.LogCall(n, nil)
	logging.InfoLog.Println("Adding commit list")
	acNode := newAllCommitsNode(n.repo)
	child := n.NewPersistentInode(ctx, acNode, fs.StableAttr{Mode: fuse.S_IFDIR})
	n.AddChild("commits", child, false)

	logging.InfoLog.Println("Adding branch list")
	blNode := newBranchListNode(n.repo)
	child = n.NewPersistentInode(ctx, blNode, fs.StableAttr{Mode: fuse.S_IFDIR})
	n.AddChild("branches", child, false)
}

func NewRootNode(path string) (node *RootNode, err error) {
	Init()
	repo, err := git.PlainOpen(path)
	if err != nil {
		return
	}

	node = &RootNode{}
	node.repo = repo
	return
}

var _ fs.NodeOnAdder = (*RootNode)(nil)
var _ fs.NodeGetattrer = (*RootNode)(nil)
