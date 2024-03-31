package gitfs

import (
	"fmt"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
	"gogitfs/pkg/logging"
	"path"
	"testing"
	"time"
)

// commitSignatures contains example commits
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
	"new": {
		Name:  "New Commiter",
		Email: "new.commiter@git.com",
		When:  time.Date(2024, 2, 7, 9, 38, 10, 0, time.UTC),
	},
}

type fileDesc struct {
	path     string
	contents string
}

// commitExtraFiles contains extra files to be created or overwritten during commits
var commitExtraFiles = map[string][]fileDesc{
	"foo": {
		{"toOverwrite.txt", "foo"},
	},
	"bar": {
		{"barDir/bar1.txt", "bar inside dir"},
		{"barDir/bar2.txt", "bar inside dir again"},
		{"toOverwrite.txt", "bar"},
	},
	"baz": {
		{"toOverwrite.txt", "baz"},
	},
}

func addFile(t *testing.T, worktree *git.Worktree, fs billy.Filesystem, desc fileDesc) error {
	err := fs.MkdirAll(path.Dir(desc.path), 0700)
	if err != nil {
		return fmt.Errorf("cannot create directories for %v: %w", desc.path, err)
	}
	f, err := fs.Create(desc.path)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return fmt.Errorf("cannot create file %v: %w", desc.path, err)
	}
	_, err = f.Write([]byte(desc.contents))
	if err != nil {
		return fmt.Errorf("cannot write to file %v: %w", desc.path, err)
	}
	_, err = worktree.Add(desc.path)
	if err != nil {
		return fmt.Errorf("cannot add file %v to worktree: %w", desc.path, err)
	}
	return nil
}

// addCommit adds a commit to worktree creating a file with name and contents equal to `msg`
// and sets the commit message to `msg`
func addCommit(t *testing.T, worktree *git.Worktree, fs billy.Filesystem, msg string) plumbing.Hash {
	errHandler := func(err error) {
		t.Fatalf("Error during creation of commit '%v': %v", msg, err)
	}
	files := append(commitExtraFiles[msg], fileDesc{path: msg, contents: msg})
	for _, f := range files {
		err := addFile(t, worktree, fs, f)
		if err != nil {
			errHandler(err)
		}
	}
	sig := commitSignatures[msg]
	opts := git.CommitOptions{Author: &sig}
	hash, err := worktree.Commit(msg, &opts)
	if err != nil {
		errHandler(fmt.Errorf("cannot commit: %w", err))
	}
	return hash
}

// checkout with error handling
func checkout(t *testing.T, worktree *git.Worktree, opts *git.CheckoutOptions) {
	err := worktree.Checkout(opts)
	if err != nil {
		t.Fatalf("Error during checkout: %v", err)
	}
}

// repoExtras represents additional information about the sample repository
type repoExtras struct {
	// commit hashes, with the same keys as commitSignatures
	commits map[string]plumbing.Hash
	// created Worktree object
	worktree *git.Worktree
	// crated Filesystem object
	fs billy.Filesystem
}

// makeRepo creates a sample repository and returns it as well as additional information
func makeRepo(t *testing.T) (repo *git.Repository, extras repoExtras) {
	logging.Init(logging.Debug)
	errHandler := func(err error) {
		t.Fatalf("Error during repository creation: %v", err)
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
	commits := make(map[string]plumbing.Hash)
	commits["foo"] = addCommit(t, worktree, fs, "foo")
	commits["bar"] = addCommit(t, worktree, fs, "bar")
	opts := git.CheckoutOptions{
		Hash:   commits["foo"],
		Branch: plumbing.NewBranchReferenceName("branch"),
		Create: true,
	}
	checkout(t, worktree, &opts)
	commits["baz"] = addCommit(t, worktree, fs, "baz")
	opts = git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("main"),
	}
	checkout(t, worktree, &opts)
	extras.commits = commits
	extras.worktree = worktree
	extras.fs = fs
	return
}

func Test_headCommit(t *testing.T) {
	repo, extras := makeRepo(t)
	n := &repoNode{repo: repo}

	c, err := headCommit(n)
	assert.NoError(t, err, "unexpected error in headCommit")
	assert.Equal(t, extras.commits["bar"], c.Hash,
		"retrieved commit's hash does not match the hash of expected head commit")
}

func Test_headAttr(t *testing.T) {
	repo, _ := makeRepo(t)
	n := &repoNode{repo: repo}

	attr, err := headAttr(n)
	assert.NoError(t, err, "unexpected error in headAttr")
	assert.Equal(t, attr.Atime, uint64(commitSignatures["bar"].When.Unix()), "incorrect access time")
	assert.Equal(t, attr.Ctime, uint64(commitSignatures["bar"].When.Unix()), "incorrect creation time")
	assert.Equal(t, attr.Mtime, uint64(commitSignatures["bar"].When.Unix()), "incorrect modification time")
}
