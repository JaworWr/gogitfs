package gitfs

import (
	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func commitNodeTestCase(t *testing.T, repo *git.Repository, extras repoExtras, commit string, hasParent bool) {
	node := &commitNode{}
	node.repo = repo
	commitObj, err := repo.CommitObject(extras.commits[commit])
	if err != nil {
		t.Fatalf("Error during commit retrieval: %v", err)
	}
	node.commit = commitObj
	server, mountPath := mountNode(t, node, noOpCb)
	defer func() {
		_ = server.Unmount()
	}()

	var children []string
	if hasParent {
		children = []string{"message", "hash", "log", "parent"}
	} else {
		children = []string{"message", "hash", "log"}
	}
	t.Run("ls", func(t *testing.T) {
		assertDirEntries(t, mountPath, children)
	})
	t.Run("stat", func(t *testing.T) {
		stat, err := os.Stat(mountPath)
		assert.NoError(t, err, "unexpected error for commit node")
		assert.Equal(t, commitSignatures[commit].When, stat.ModTime().UTC(), "incorrect time for commit node")

		for _, c := range []string{"message", "hash"} {
			stat, err := os.Stat(path.Join(mountPath, c))
			assert.NoError(t, err, "unexpected error for %v", c)
			assert.Equal(t, commitSignatures[commit].When, stat.ModTime().UTC(), "incorrect time for %v", c)
		}
	})
}

func Test_commitNode(t *testing.T) {
	repo, extras := makeRepo(t)
	testCases := []struct {
		name, commit string
		hasParent    bool
	}{
		{
			"head commit",
			"bar",
			true,
		},
		{
			"last commit",
			"foo",
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			commitNodeTestCase(t, repo, extras, tc.commit, tc.hasParent)
		})
	}
}
