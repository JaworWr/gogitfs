package gitfs

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_branchListNode(t *testing.T) {
	Init()
	repo, extras := makeRepo(t)
	node := newBranchListNode(repo)
	server, mountPath := mountNode(t, node, noOpCb)
	defer func() {
		_ = server.Unmount()
	}()

	expected := []string{"main", "branch"}
	t.Run("ls", func(t *testing.T) {
		assertDirEntries(t, mountPath, expected, "unexpected ls result")
	})
	t.Run("stat", func(t *testing.T) {
		stat, err := os.Stat(mountPath)
		assert.NoError(t, err)
		assert.Equal(t, commitSignatures["bar"].When, stat.ModTime().UTC())
	})

	opts := git.CheckoutOptions{
		Hash:   extras.commits["foo"],
		Branch: plumbing.NewBranchReferenceName("branch2"),
		Create: true,
	}
	checkout(t, extras.worktree, &opts)
	expected = append(expected, "branch2")
	t.Run("ls with added branch", func(t *testing.T) {
		assertDirEntries(t, mountPath, expected, "unexpected ls result")
	})
}
