package gitfs

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"slices"
	"strings"
	"testing"
)

type expectedDirEntry struct {
	name string
	mode os.FileMode
}

func dirNodeTestCase(
	t *testing.T,
	repo *git.Repository,
	extras repoExtras,
	commit string,
	dirPath string,
	expected []expectedDirEntry,
) {
	var hash plumbing.Hash
	if commit == "HEAD" {
		head, err := repo.Head()
		if err != nil {
			t.Fatalf("Cannot get repo head: %v", err)
		}
		hash = head.Hash()
	} else {
		hash = extras.commits[commit]
	}
	commitObj, err := repo.CommitObject(hash)
	if err != nil {
		t.Fatalf("Error during commit retrieval: %v", err)
	}
	node, err := newTreeDirNode(repo, commitObj)
	if err != nil {
		t.Fatalf("Error during node creation: %v", err)
	}
	server, mountPath := mountNode(t, node, fs.StableAttr{Mode: fuse.S_IFDIR}, noOpCb)
	defer func(server *fuse.Server) {
		_ = server.Unmount()
	}(server)

	slices.SortFunc(expected, func(a, b expectedDirEntry) int {
		return strings.Compare(a.name, b.name)
	})
	entries, err := os.ReadDir(path.Join(mountPath, dirPath))
	assert.NoError(t, err, "unexpected error when reading directory")
	assert.Equal(t, len(expected), len(entries), "result length mismatch")
	for i, entry := range entries {
		expectedEntry := expected[i]
		assert.Equal(t, expectedEntry.name, entry.Name(), "name mismatch")
		assert.Equal(t, expectedEntry.mode.Type(), entry.Type(), "type mismatch")
		info, err := entry.Info()
		if err != nil {
			t.Fatalf("cannot get entry info: %v", err)
		}
		assert.Equal(t, expectedEntry.mode, info.Mode(), "mode mismatch")
		assert.Equal(t, commitSignatures[commit].When, info.ModTime().UTC(), "modification mode mismatch")
	}
}

func Test_dirNode(t *testing.T) {
	repo, extras := makeRepo(t)
	testCases := []struct {
		commit, dirPath string
		expected        []expectedDirEntry
	}{
		{"bar", ".", []expectedDirEntry{
			{"toOverwrite.txt", 0444},
			{"barDir", os.ModeDir | 0555},
		}},
	}

	for _, tc := range testCases {
		name := fmt.Sprintf("%v:(%v)", tc.commit, tc.dirPath)
		t.Run(name, func(t *testing.T) {
			dirNodeTestCase(t, repo, extras, tc.commit, tc.dirPath, tc.expected)
		})
	}
}
