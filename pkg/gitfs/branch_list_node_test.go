package gitfs

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
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
		assertDirEntries(t, mountPath, expected, "incorrect directory entries")
	})
	t.Run("stat", func(t *testing.T) {
		stat, err := os.Stat(mountPath)
		assert.NoError(t, err, "unexpected error on running os.Stat")
		assert.Equal(t, commitSignatures["bar"].When, stat.ModTime().UTC(), "incorrect modification time")
	})

	opts := git.CheckoutOptions{
		Hash:   extras.commits["foo"],
		Branch: plumbing.NewBranchReferenceName("branch2"),
		Create: true,
	}
	checkout(t, extras.worktree, &opts)
	expected = append(expected, "branch2")
	t.Run("ls with added branch", func(t *testing.T) {
		assertDirEntries(t, mountPath, expected, "incorrect directory entries")
	})

	t.Run("lookup existent", func(t *testing.T) {
		_, err := os.Stat(path.Join(mountPath, "branch"))
		assert.NoError(t, err, "unexpected error on running os.Stat on existent branch's node")
	})

	t.Run("lookup nonexistent", func(t *testing.T) {
		_, err := os.Stat(path.Join(mountPath, "nonexistent"))
		assert.Error(t, err, "expected an error on running os.Stat on nonexistent branch's node")
		assert.True(t, os.IsNotExist(err), "error should be an ErrNotExist")
	})
}
