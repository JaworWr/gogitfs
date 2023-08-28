package gitfs

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
	"gogitfs/pkg/logging"
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
	_, err = worktree.Add(msg)
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

func makeRepo(t *testing.T) (repo *git.Repository, commits map[string]plumbing.Hash) {
	logging.Init(logging.Debug)
	errHandler := func(err error) {
		t.Fatalf("Error during repo creation: %v", err)
	}
	tempdir := t.TempDir()
	fs := osfs.New(tempdir)
	storage := memory.NewStorage()
	initOpts := git.InitOptions{DefaultBranch: plumbing.NewBranchReferenceName("main")}
	repo, err := git.InitWithOptions(storage, fs, initOpts)
	if err != nil {
		errHandler(err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		errHandler(err)
	}
	commits = make(map[string]plumbing.Hash)
	commits["foo"] = addCommit(t, worktree, fs, "foo")
	commits["bar"] = addCommit(t, worktree, fs, "bar")
	opts := git.CheckoutOptions{
		Hash:   commits["foo"],
		Branch: plumbing.NewBranchReferenceName("branch"),
		Create: true,
	}
	err = worktree.Checkout(&opts)
	if err != nil {
		errHandler(err)
	}
	commits["baz"] = addCommit(t, worktree, fs, "baz")
	opts = git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("main"),
	}
	err = worktree.Checkout(&opts)
	if err != nil {
		errHandler(err)
	}
	return
}

func Test_headCommit(t *testing.T) {
	repo, commits := makeRepo(t)
	n := &repoNode{repo: repo}

	c, err := headCommit(n)
	assert.NoError(t, err)
	assert.Equal(t, commits["bar"], c.Hash)
}

func Test_headAttr(t *testing.T) {
	repo, _ := makeRepo(t)
	n := &repoNode{repo: repo}

	attr, err := headAttr(n)
	assert.NoError(t, err)
	assert.Equal(t, attr.Atime, uint64(commitSignatures["bar"].When.Unix()))
	assert.Equal(t, attr.Ctime, uint64(commitSignatures["bar"].When.Unix()))
	assert.Equal(t, attr.Mtime, uint64(commitSignatures["bar"].When.Unix()))
}