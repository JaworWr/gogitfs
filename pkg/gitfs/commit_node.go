package gitfs

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/error_handler"
	"gogitfs/pkg/gitfs/internal/utils"
	"gogitfs/pkg/logging"
	"strings"
	"syscall"
	"time"
)

type commitNode struct {
	repoNode
	commit *object.Commit
}

func (n *commitNode) GetCallCtx() logging.CallCtx {
	info := utils.NodeCallCtx(n)
	info["hash"] = n.commit.Hash.String()
	info["msg"] = n.commit.Message
	return info
}

func (n *commitNode) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	logging.LogCall(n, nil)
	out.Attr = utils.CommitAttr(n.commit)
	out.Mode = 0555
	return 0
}

func (n *commitNode) OnAdd(ctx context.Context) {
	defer logging.Benchmark(time.Now())
	logging.LogCall(n, nil)
	attr := utils.CommitAttr(n.commit)
	attr.Mode = 0444
	hashNode := &fs.MemRegularFile{Attr: attr, Data: []byte(n.commit.Hash.String())}
	child := n.NewPersistentInode(ctx, hashNode, fs.StableAttr{Mode: fuse.S_IFREG})
	n.AddChild("hash", child, false)

	msgNode := &fs.MemRegularFile{Attr: attr, Data: []byte(n.commit.Message)}
	child = n.NewPersistentInode(ctx, msgNode, fs.StableAttr{Mode: fuse.S_IFREG})
	n.AddChild("message", child, false)

	parent, err := n.commit.Parent(0)
	if err == nil {
		parentAttr := utils.CommitAttr(parent)
		parentAttr.Mode = 0555
		path := fmt.Sprintf("../%v", parent.Hash.String())
		parentNode := &fs.MemSymlink{Attr: parentAttr, Data: []byte(path)}
		child = n.NewPersistentInode(ctx, parentNode, fs.StableAttr{Mode: fuse.S_IFLNK})
		n.AddChild("parent", child, false)
	} else if err != object.ErrParentNotFound {
		error_handler.Fatal.HandleError(err)
	}

	nodeOpts := commitLogNodeOpts{linkLevels: 2}
	logNode, err := newCommitLogNode(n.repo, n.commit, nodeOpts)
	if err != nil {
		error_handler.Fatal.HandleError(err)
	}
	child = n.NewPersistentInode(ctx, logNode, fs.StableAttr{Mode: fuse.S_IFDIR})
	n.AddChild("log", child, false)
}

func newCommitNode(ctx context.Context, commit *object.Commit, parent repoNodeEmbedder) *fs.Inode {
	builder := func() (fs.InodeEmbedder, error) {
		logging.InfoLog.Printf(
			"Creating new node for commit %v (%v)",
			commit.Hash,
			strings.Replace(commit.Message, "\n", ";", -1),
		)
		node := commitNode{commit: commit}
		node.repo = parent.embeddedRepoNode().repo
		return &node, nil
	}
	node, _ := commitNodeMgr.GetOrInsert(ctx, commit.Hash.String(), fuse.S_IFDIR, parent, builder, false)
	return node
}

var _ fs.NodeOnAdder = (*commitNode)(nil)
var _ fs.NodeGetattrer = (*commitNode)(nil)
