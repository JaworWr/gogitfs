package gitfs

import (
	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func catFile(t *testing.T, path string) string {
	data, err := os.ReadFile(path)
	assert.NoError(t, err, "unexpected error when reading %v", path)
	return string(data)
}

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
		children = []string{"message", "hash", "log", "parent", "parents"}
	} else {
		children = []string{"message", "hash", "log", "parents"}
	}
	t.Run("ls", func(t *testing.T) {
		assertDirEntries(t, mountPath, children)
	})
	t.Run("stat", func(t *testing.T) {
		stat, err := os.Stat(mountPath)
		assert.NoError(t, err, "unexpected error for commit node")
		assert.Equal(t, commitSignatures[commit].When, stat.ModTime().UTC(), "incorrect time for commit node")

		for _, c := range []string{"message", "hash"} {
			t.Run(c, func(t *testing.T) {
				stat, err := os.Stat(path.Join(mountPath, c))
				assert.NoError(t, err, "unexpected error")
				assert.Equal(t, commitSignatures[commit].When, stat.ModTime().UTC(), "incorrect time")
			})
		}
	})
	t.Run("cat", func(t *testing.T) {
		t.Run("message", func(t *testing.T) {
			result := catFile(t, path.Join(mountPath, "message"))
			assert.Equal(t, commit, result)
		})
		t.Run("hash", func(t *testing.T) {
			result := catFile(t, path.Join(mountPath, "hash"))
			assert.Equal(t, extras.commits[commit].String(), result)
		})
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
