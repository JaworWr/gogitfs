package gitfs

import (
	"github.com/go-git/go-git/v5"
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

	var expected []string
	if hasParent {
		expected = []string{"message", "hash", "log", "parent"}
	} else {
		expected = []string{"message", "hash", "log"}
	}
	t.Run("ls", func(t *testing.T) {
		assertDirEntries(t, mountPath, expected)
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
