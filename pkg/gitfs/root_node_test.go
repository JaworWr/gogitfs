package gitfs

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"sort"
	"testing"
)

type MountCb = func(t *testing.T, ctx context.Context, inode *fs.Inode)

func noOpCb(_ *testing.T, _ context.Context, _ *fs.Inode) {

}

func mountNode(t *testing.T, n fs.InodeEmbedder, cb MountCb) (server *fuse.Server, mountPath string) {
	tmpdir := t.TempDir()
	mountPath = path.Join(tmpdir, "root")
	root := &fs.Inode{}
	opts := &fs.Options{}
	opts.OnAdd = func(ctx context.Context) {
		node := root.NewPersistentInode(ctx, n, fs.StableAttr{Mode: fuse.S_IFDIR})
		root.AddChild("root", node, false)
		cb(t, ctx, node)
	}
	server, err := fs.Mount(tmpdir, root, opts)
	if err != nil {
		t.Fatalf("Cannot mount server. Error: %v", err)
	}
	return
}

func getSortedNames(entries []os.DirEntry) []string {
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	sort.Strings(names)
	return names
}

func Test_RootNode(t *testing.T) {
	node := &RootNode{}
	repo, _ := makeRepo(t)
	node.repo = repo
	server, mountPath := mountNode(t, node, noOpCb)
	defer func() {
		_ = server.Unmount()
	}()
	entries, err := os.ReadDir(mountPath)
	names := getSortedNames(entries)
	assert.NoError(t, err)
	assert.Equal(t, []string{"branches", "commits"}, names)
	stat, err := os.Stat(mountPath)
	assert.NoError(t, err)
	assert.Equal(t, commitSignatures["bar"].When, stat.ModTime().UTC())
}
