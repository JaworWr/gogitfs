package gitfs

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
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

			p := string(node.Data)
			assert.Equal(t, tc.expectedPrefix+hash.String(), p, "incorrect path")

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

	expectedCommits := make([]string, len(expected.commits))
	for i, c := range expected.commits {
		expectedCommits[i] = extras.commits[c].String()
	}
	t.Run("ls", func(t *testing.T) {
		expectedEntries := expectedCommits
		if expected.expectHeadLink {
			expectedEntries = append(expectedEntries, "HEAD")
		}
		assertDirEntries(t, mountPath, expectedEntries, "incorrect directory entries")
	})

	t.Run("head link", func(t *testing.T) {
		if !expected.expectHeadLink {
			return
		}
		p, err := os.Readlink(path.Join(mountPath, "HEAD"))
		assert.NoError(t, err, "unexpected Readlink error")
		assert.Equal(t, expected.headLink, p)
	})

	t.Run("symlinks", func(t *testing.T) {
		p := path.Join(mountPath, expectedCommits[0])
		stat, err := os.Lstat(p)
		assert.NoError(t, err, "unexpected Stat error")
		if expected.expectSymlinks {
			assert.Equal(t, os.ModeSymlink, stat.Mode()&os.ModeSymlink, "commit node should be a symlink")
			p1, err := os.Readlink(p)
			assert.NoError(t, err, "unexpected Readlink error")
			assert.Equal(t, expected.symlinkPrefix+expectedCommits[0], p1, "incorrect symlink path")
		} else {
			assert.Equal(t, os.ModeDir, stat.Mode()&os.ModeDir, "commit node should be a directory")
		}
	})
}

func Test_CommitLogNode(t *testing.T) {
	Init()
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
		{
			"from bar",
			args{"bar", commitLogNodeOpts{0, true, false}},
			commitLogNodeTestExpected{
				commits:        []string{"bar", "foo"},
				expectHeadLink: false,
				expectSymlinks: false,
			},
		},
		{
			"from baz",
			args{"baz", commitLogNodeOpts{0, true, false}},
			commitLogNodeTestExpected{
				commits:        []string{"baz", "foo"},
				expectHeadLink: false,
				expectSymlinks: false,
			},
		},
		{
			"no head",
			args{"bar", commitLogNodeOpts{0, false, false}},
			commitLogNodeTestExpected{
				commits:        []string{"foo"},
				expectHeadLink: false,
				expectSymlinks: false,
			},
		},
		{
			"head symlink",
			args{"bar", commitLogNodeOpts{0, true, true}},
			commitLogNodeTestExpected{
				commits:        []string{"bar", "foo"},
				expectHeadLink: true,
				headLink:       extras.commits["bar"].String(),
				expectSymlinks: false,
			},
		},
		{
			"symlinks",
			args{"bar", commitLogNodeOpts{2, true, false}},
			commitLogNodeTestExpected{
				commits:        []string{"bar", "foo"},
				expectHeadLink: false,
				expectSymlinks: true,
				symlinkPrefix:  "../../",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			commitObj, err := repo.CommitObject(extras.commits[tc.from])
			if err != nil {
				t.Fatalf("Error during commit retrieval: %v", err)
			}
			node, err := newCommitLogNode(repo, commitObj, tc.opts)
			assert.NoError(t, err, "unexpected error during node creation")
			commitLogNodeTestCase(t, extras, node, tc.expected)
		})
	}
}
