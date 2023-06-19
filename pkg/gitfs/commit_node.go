package gitfs

import (
	"context"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"syscall"
)

type commitNode struct {
	repoNode
	commit *object.Commit
}

func (n *commitNode) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Attr = n.attr()
	out.Mode = 0555
	return 0
}

func (n *commitNode) OnAdd(ctx context.Context) {
	attr := n.attr()
	attr.Mode = 0444
	hashNode := &fs.MemRegularFile{Attr: attr, Data: []byte(n.commit.Hash.String())}
	child := n.NewPersistentInode(ctx, hashNode, fs.StableAttr{Mode: fuse.S_IFREG})
	n.AddChild("hash", child, false)

	msgNode := &fs.MemRegularFile{Attr: attr, Data: []byte(n.commit.Message)}
	child = n.NewPersistentInode(ctx, msgNode, fs.StableAttr{Mode: fuse.S_IFREG})
	n.AddChild("message", child, false)
}

func newCommitNode(ctx context.Context, commit *object.Commit, parent repoNodeEmbedder) *fs.Inode {
	builder := func() fs.InodeEmbedder {
		node := commitNode{commit: commit}
		node.repo = parent.embeddedRepoNode().repo
		return &node
	}
	return commitNodeMgr.GetOrInsert(ctx, commit.Hash, fuse.S_IFDIR, parent, builder, false)
}

func (n *commitNode) attr() fuse.Attr {
	commitTime := (uint64)(n.commit.Author.When.Unix())
	return fuse.Attr{
		Atime: commitTime,
		Ctime: commitTime,
		Mtime: commitTime,
	}
}

var _ fs.NodeOnAdder = (*commitNode)(nil)
var _ fs.NodeGetattrer = (*commitNode)(nil)
