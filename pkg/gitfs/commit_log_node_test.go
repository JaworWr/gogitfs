package gitfs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_commitSymlink(t *testing.T) {
	repo, extras := makeRepo(t)
	basePath := "asdf"
	type args struct {
		commit   string
		basePath *string
	}
	testCases := []struct {
		name string
		args
		expectedPrefix string
	}{
		{
			"foo without prefix",
			args{"foo", nil},
			"",
		},
		{
			"bar with prefix",
			args{"bar", &basePath},
			"asdf/",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash := extras.commits[tc.commit]
			commit, err := repo.CommitObject(hash)
			if err != nil {
				t.Fatalf("Error during retrieval of commit object: %v", err)
			}
			node := commitSymlink(commit, tc.basePath)

			path := string(node.Data)
			assert.Equal(t, tc.expectedPrefix+hash.String(), path, "incorrect path")

			commitTime := commitSignatures[tc.commit].When.Unix()
			assert.EqualValues(t, commitTime, node.Attr.Mtime, "incorrect mtime")
			assert.EqualValues(t, commitTime, node.Attr.Atime, "incorrect atime")
			assert.EqualValues(t, commitTime, node.Attr.Ctime, "incorrect ctime")
		})
	}
}

func Test_getBasePath(t *testing.T) {
	tests := []struct {
		name       string
		linkLevels int
		expectNil  bool
		expected   string
	}{
		{
			"zero",
			0,
			true,
			"",
		},
		{
			"non-zero",
			2,
			false,
			"../..",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			basePath := getBasePath(tc.linkLevels)
			if tc.expectNil {
				assert.Nil(t, basePath)
			} else {
				assert.NotNil(t, basePath)
				assert.Equal(t, tc.expected, *basePath)
			}
		})
	}
}

type commitLogNodeTestExpected struct {
	commits        []string
	expectHead     bool
	expectHeadLink bool
	headLink       string
	expectSymlinks bool
	symlinkPrefix  string
}

func commitLogNodeTestCase(t *testing.T, extras repoExtras, node *commitLogNode, expected commitLogNodeTestExpected) {
	server, mountPath := mountNode(t, node, noOpCb)
	defer func() {
		_ = server.Unmount()
	}()
	_ = mountPath
}

func Test_CommitLogNode(t *testing.T) {
	repo, extras := makeRepo(t)
	type args struct {
		from string
		opts commitLogNodeOpts
	}
	testCases := []struct {
		name string
		args
		expected commitLogNodeTestExpected
	}{
		// TODO cases
	}
	for _, tc := range testCases {
		commitObj, err := repo.CommitObject(extras.commits[tc.from])
		if err != nil {
			t.Fatalf("Error during commit retrieval: %v", err)
		}
		node, err := newCommitLogNode(repo, commitObj, tc.opts)
		assert.NoError(t, err, "unexpected error during node creation")
		commitLogNodeTestCase(t, extras, node, tc.expected)
	}
}
