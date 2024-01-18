package gitfs

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func Test_allCommitsNode(t *testing.T) {
	Init()
	repo, extras := makeRepo(t)
	node := newAllCommitsNode(repo)
	server, mountPath := mountNode(t, node, noOpCb)
	defer func() {
		_ = server.Unmount()
	}()

	expected := []string{"HEAD"}
	for _, hash := range extras.commits {
		expected = append(expected, hash.String())
	}
	t.Run("ls", func(t *testing.T) {
		assertDirEntries(t, mountPath, expected, "incorrect directory entries")
	})

	t.Run("readlink", func(t *testing.T) {
		p, err := os.Readlink(path.Join(mountPath, "HEAD"))
		assert.NoError(t, err, "unexpected error when reading HEAD symlink")
		assert.Equal(t, extras.commits["bar"].String(), p, "incorrect HEAD symlink path")
	})

	t.Run("stat", func(t *testing.T) {
		stat, err := os.Stat(mountPath)
		assert.NoError(t, err, "unexpected error on running os.Stat")
		assert.Equal(t, commitSignatures["bar"].When, stat.ModTime().UTC(), "incorrect modification time")
	})

	t.Run("HEAD stat", func(t *testing.T) {
		stat, err := os.Lstat(path.Join(mountPath, "HEAD"))
		assert.NoError(t, err, "unexpected error on running os.Lstat")
		assert.Equal(t, commitSignatures["bar"].When, stat.ModTime().UTC(), "incorrect modification time")
	})

	hash := addCommit(t, extras.worktree, extras.fs, "new")
	expected = append(expected, hash.String())

	t.Run("ls with added commit", func(t *testing.T) {
		assertDirEntries(t, mountPath, expected, "incorrect directory entries")
	})

	t.Run("stat with added commit", func(t *testing.T) {
		stat, err := os.Stat(mountPath)
		assert.NoError(t, err, "unexpected error on running os.Stat")
		assert.Equal(t, commitSignatures["new"].When, stat.ModTime().UTC(), "incorrect modification time")
	})
}
