package gitfs

import (
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/stretchr/testify/assert"
	"os"
	"sort"
	"testing"
)

func mountNode(t *testing.T, n fs.InodeEmbedder) (server *fuse.Server, path string) {
	path = t.TempDir()
	server, err := fs.Mount(path, n, &fs.Options{})
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
	server, path := mountNode(t, node)
	defer func() {
		_ = server.Unmount()
	}()
	entries, err := os.ReadDir(path)
	names := getSortedNames(entries)
	assert.NoError(t, err)
	assert.Equal(t, []string{"branches", "commits"}, names)

}
