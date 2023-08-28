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
	assertDirEntries(t, mountPath, expected, "unexpected ls result")

	//hash = add
}
