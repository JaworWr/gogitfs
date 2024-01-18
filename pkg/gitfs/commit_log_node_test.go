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
			"symlink without prefix",
			args{"foo", nil},
			"",
		},
		{
			"symlink with prefix",
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
			assert.Equal(t, tc.expectedPrefix+hash.String(), p, "incorrect symlink path")

			commitTime := commitSignatures[tc.commit].When.Unix()
			assert.EqualValues(t, commitTime, node.Attr.Mtime, "incorrect symlink mtime")
			assert.EqualValues(t, commitTime, node.Attr.Atime, "incorrect symlink atime")
			assert.EqualValues(t, commitTime, node.Attr.Ctime, "incorrect symlink ctime")
		})
	}
}

func Test_getBasePath(t *testing.T) {
	expectedBasePath := "../.."
	tests := []struct {
		name       string
		linkLevels int
		expected   *string
	}{
		{
			"zero",
			0,
			nil,
		},
		{
			"non-zero",
			2,
			&expectedBasePath,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			basePath := getBasePath(tc.linkLevels)
			assert.Equal(t, tc.expected, basePath, "incorrect basepath")
		})
	}
}

// commitLogNodeTestExpected is the expected result of a single commitLogNode test case
type commitLogNodeTestExpected struct {
	// expected commit hashes
	commits []string
	// whether there should be a HEAD symlink
	expectHeadLink bool
	// where the HEAD symlink should point (if present)
	headLink string
	// whether the commit node should be a symlink
	expectSymlinks bool
	// prefix of the symlink, if the node is one
	symlinkPrefix string
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

	t.Run("HEAD symlink", func(t *testing.T) {
		if !expected.expectHeadLink {
			return
		}
		p, err := os.Readlink(path.Join(mountPath, "HEAD"))
		assert.NoError(t, err, "unexpected Readlink error")
		assert.Equal(t, expected.headLink, p, "incorrect HEAD symlink path")
	})

	t.Run("symlinks", func(t *testing.T) {
		for _, c := range expectedCommits {
			p := path.Join(mountPath, c)
			stat, err := os.Lstat(p)
			assert.NoError(t, err, "unexpected os.Lstat error")
			if expected.expectSymlinks {
				assert.Equal(t, os.ModeSymlink, stat.Mode()&os.ModeSymlink, "commit node should be a symlink")
				p1, err := os.Readlink(p)
				assert.NoError(t, err, "unexpected Readlink error")
				assert.Equal(t, expected.symlinkPrefix+c, p1, "incorrect symlink path")
			} else {
				assert.Equal(t, os.ModeDir, stat.Mode()&os.ModeDir, "commit node should be a directory")
			}
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
			"no HEAD",
			args{"bar", commitLogNodeOpts{0, false, false}},
			commitLogNodeTestExpected{
				commits:        []string{"foo"},
				expectHeadLink: false,
				expectSymlinks: false,
			},
		},
		{
			"HEAD symlink",
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
