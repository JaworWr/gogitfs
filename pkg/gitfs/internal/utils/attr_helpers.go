package utils

import (
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func CommitAttr(commit *object.Commit) fuse.Attr {
	commitTime := (uint64)(commit.Author.When.Unix())
	return fuse.Attr{
		Atime: commitTime,
		Ctime: commitTime,
		Mtime: commitTime,
	}
}
