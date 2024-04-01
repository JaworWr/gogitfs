package gitfs

import (
	"bytes"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/stretchr/testify/assert"
	"gogitfs/pkg/gitfs/internal/utils"
	"os"
	"testing"
)

// getFile retrieves a file in a commit's tree by path
func getFile(repo *git.Repository, commitObj *object.Commit, path string) (*object.File, error) {
	tree, err := repo.TreeObject(commitObj.TreeHash)
	if err != nil {
		return nil, fmt.Errorf("cannot get tree object: %w", err)
	}
	file, err := tree.File(path)
	if err != nil {
		return nil, fmt.Errorf("cannot get file: %w", err)
	}
	return file, nil
}

func fileNodeTestCase(
	t *testing.T,
	repo *git.Repository,
	extras repoExtras,
	commit string,
	fileCommit string,
	path string,
) {
	var hash plumbing.Hash
	if commit == "HEAD" {
		head, err := repo.Head()
		if err != nil {
			t.Fatalf("Cannot get repo head: %v", err)
		}
		hash = head.Hash()
	} else {
		hash = extras.commits[commit]
	}
	commitObj, err := repo.CommitObject(hash)
	if err != nil {
		t.Fatalf("Error during commit retrieval: %v", err)
	}
	file, err := getFile(repo, commitObj, path)
	if err != nil {
		t.Fatalf("Error during file retrieval: %v", err)
	}
	attr := utils.CommitAttr(commitObj)
	node := newFileNode(file, attr)
	server, mountPath := mountNode(t, node, fs.StableAttr{Mode: fuse.S_IFREG}, noOpCb)
	defer func(server *fuse.Server) {
		_ = server.Unmount()
	}(server)

	t.Run("read", func(t *testing.T) {
		f, err := os.Open(mountPath)
		defer func() { _ = f.Close() }()
		assert.NoError(t, err, "unexpected error when opening file")
		buf := &bytes.Buffer{}
		_, err = buf.ReadFrom(f)
		assert.NoError(t, err, "unexpected error when reading file")
		contents, ok := commitExtraFiles[fileCommit][path]
		if !ok {
			contents = path
		}
		assert.Equal(t, contents, buf.String())
	})
	t.Run("stat", func(t *testing.T) {
		stat, err := os.Stat(mountPath)
		assert.NoError(t, err, "unexpected error when calling Stat()")
		assert.Equal(t, node.file.Size, stat.Size(), "size mismatch")
		assert.Equal(t, os.FileMode(0444), stat.Mode().Perm(), "mode mismatch")
		if commit != "HEAD" {
			assert.Equal(t, commitSignatures[commit].When, stat.ModTime().UTC(), "modification time mismatch")
		}
	})
}

func Test_fileNode(t *testing.T) {
	repo, extras := makeRepo(t)
	testCases := []struct {
		commit, fileCommit, path string
	}{
		{"HEAD", "foo", "foo"},
		{"HEAD", "bar", "barDir/bar1.txt"},
		{"HEAD", "bar", "barDir/bar2.txt"},
		{"foo", "foo", "toOverwrite.txt"},
		{"bar", "bar", "toOverwrite.txt"},
		{"baz", "baz", "toOverwrite.txt"},
		{"HEAD", "bar", "toOverwrite.txt"},
	}

	for _, tc := range testCases {
		name := fmt.Sprintf("%v:(%v)", tc.commit, tc.path)
		t.Run(name, func(t *testing.T) {
			fileNodeTestCase(t, repo, extras, tc.commit, tc.fileCommit, tc.path)
		})
	}
}
