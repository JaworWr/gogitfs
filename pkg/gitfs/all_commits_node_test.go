package gitfs

import "testing"

func Test_allCommitsNode(t *testing.T) {
	Init()
	repo, commits := makeRepo(t)
	node := newAllCommitsNode(repo)
	server, mountPath := mountNode(t, node, noOpCb)
	defer func() {
		_ = server.Unmount()
	}()

	expected := []string{"HEAD"}
	for _, hash := range commits {
		expected = append(expected, hash.String())
	}
	assertDirEntries(t, mountPath, expected, "unexpected ls result")
}
