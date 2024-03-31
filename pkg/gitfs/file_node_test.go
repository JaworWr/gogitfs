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

func getFile(repo *git.Repository, commit plumbing.Hash, path string) (*object.File, error) {
	commitObj, err := repo.CommitObject(commit)
	if err != nil {
		return nil, fmt.Errorf("cannot get commit object: %w", err)
	}
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

func Test_FileNode(t *testing.T) {
	repo, extras := makeRepo(t)
	head, err := repo.Head()
	if err != nil {
		t.Fatalf("Cannot get repo head: %v", err)
	}
	file, err := getFile(repo, head.Hash(), "foo")
	if err != nil {
		t.Fatalf("Error during file retrieval: %v", err)
	}
	commitObj, err := repo.CommitObject(extras.commits["foo"])
	if err != nil {
		t.Fatalf("Error during commit retrieval: %v", err)
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
		assert.Equal(t, "foo", buf.String())
	})
}
