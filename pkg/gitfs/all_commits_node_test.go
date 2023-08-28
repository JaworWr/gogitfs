package gitfs

import "testing"

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
		assertDirEntries(t, mountPath, expected, "unexpected ls result")
	})

	t.Run("ls with added commit", func(t *testing.T) {
		hash := addCommit(t, extras.worktree, extras.fs, "aaa")
		expected = append(expected, hash.String())
		assertDirEntries(t, mountPath, expected, "unexpected ls result")
	})
}
