package gitfs

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"testing"
	"time"
)

var commitSignatures = map[string]object.Signature{
	"foo": {
		Name:  "Aaa Bbb",
		Email: "foo@bar.com",
		When:  time.Date(2023, 1, 10, 12, 34, 56, 0, time.UTC),
	},
	"bar": {
		Name:  "Ccc Ddd",
		Email: "ccc@ddd.com",
		When:  time.Date(2023, 2, 5, 9, 32, 12, 0, time.UTC),
	},
	"baz": {
		Name:  "Eee Fff",
		Email: "ef@ef.com",
		When:  time.Date(2023, 2, 7, 9, 32, 10, 0, time.UTC),
	},
}

func addCommit(t *testing.T, worktree *git.Worktree, fs billy.Filesystem, msg string) plumbing.Hash {
	errHandler := func(err error) {
		t.Fatalf("Error during commit creation: %v", err)
	}
	f, err := fs.Create(msg)
	if err != nil {
		errHandler(err)
	}
	_, err = f.Write([]byte(msg))
	if err != nil {
		errHandler(err)
	}
	err = f.Close()
	if err != nil {
		errHandler(err)
	}
	sig := commitSignatures[msg]
	opts := git.CommitOptions{Author: &sig}
	hash, err := worktree.Commit(msg, &opts)
	if err != nil {
		errHandler(err)
	}
	return hash
}

func makeRepo(t *testing.T) *git.Repository {
	errHandler := func(err error) {
		t.Fatalf("Error during repo creation: %v", err)
	}
	tempdir := t.TempDir()
	fs := osfs.New(tempdir)
	storage := memory.NewStorage()
	repo, err := git.Init(storage, fs)
	if err != nil {
		errHandler(err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		errHandler(err)
	}
	_ = worktree
	return repo
}
