package gitfs

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"path"
	"testing"
)

type MountCb = func(t *testing.T, ctx context.Context, path string, root *fs.Inode)

func testWithMount(t *testing.T, n fs.InodeEmbedder, cb MountCb) {
	tmpdir := t.TempDir()
	root := &fs.Inode{}
	opts := &fs.Options{}
	opts.OnAdd = func(ctx context.Context) {
		node := root.NewPersistentInode(ctx, n, fs.StableAttr{Mode: fuse.S_IFDIR})
		root.AddChild("root", node, false)
		cb(t, ctx, path.Join(tmpdir, "root"), root)
	}
	server, err := fs.Mount(tmpdir, root, opts)
	if err != nil {
		t.Fatalf("Cannot mount server. Error: %v", err)
	}
	_ = server.Unmount()
}
