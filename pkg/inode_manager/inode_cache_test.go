package inode_manager

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	"testing"
)

func testWithNode(t *testing.T, cb func(t *testing.T, ctx context.Context, inode *fs.Inode)) {
	tmpdir := t.TempDir()
	root := &fs.Inode{}
	opts := &fs.Options{}
	opts.OnAdd = func(ctx context.Context) {
		cb(t, ctx, root)
	}
	server, err := fs.Mount(tmpdir, root, opts)
	if err != nil {
		t.Fatalf("Cannot mount server. Error: %v", err)
	}
	_ = server.Unmount()
}
