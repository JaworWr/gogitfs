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

// MountCb is the type of callbacks run after mounting the node
type MountCb = func(t *testing.T, ctx context.Context, inode *fs.Inode)

func noOpCb(_ *testing.T, _ context.Context, _ *fs.Inode) {

}

// mountNode mounts the given InodeEmbedder at a temporary path and returns the mount server object
// and the path of the mounted node in the filesystem.
// The function also runs a callback after adding the node - this allows code to be run which requires
// the node to be mounted.
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

// getSortedNames extracts name from a list of directory entries and returns them in sorted order.
func getSortedNames(entries []os.DirEntry) []string {
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	sort.Strings(names)
	return names
}

// assertDirEntries checks if the directory at path contains the specified files
func assertDirEntries(t *testing.T, path string, expected []string, msgAndArgs ...interface{}) {
	entries, err := os.ReadDir(path)
	names := getSortedNames(entries)
	sorted := make([]string, len(expected))
	copy(sorted, expected)
	sort.Strings(sorted)
	assert.NoError(t, err, "unexpected error during directory reading")
	assert.Equal(t, sorted, names, msgAndArgs...)
}

func Test_RootNode(t *testing.T) {
	node := &RootNode{}
	repo, _ := makeRepo(t)
	node.repo = repo
	server, mountPath := mountNode(t, node, noOpCb)
	defer func() {
		_ = server.Unmount()
	}()
	t.Run("ls", func(t *testing.T) {
		expected := []string{"branches", "commits"}
		assertDirEntries(t, mountPath, expected, "unexpected ls result")
	})
	t.Run("stat", func(t *testing.T) {
		stat, err := os.Stat(mountPath)
		assert.NoError(t, err, "unexpected error when running os.Stat")
		assert.Equal(t, commitSignatures["bar"].When, stat.ModTime().UTC(), "incorrect modification time")
	})
}
