package utils

import (
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/logging"
)

func CommitAttr(commit *object.Commit) fuse.Attr {
	commitTime := (uint64)(commit.Author.When.Unix())
	return fuse.Attr{
		Atime: commitTime,
		Ctime: commitTime,
		Mtime: commitTime,
	}
}

func NodeCallCtx(n fs.InodeEmbedder) logging.CallCtx {
	result := make(logging.CallCtx)
	attr := n.EmbeddedInode().StableAttr()
	result["ino"] = attr.Ino
	result["gen"] = attr.Gen
	return result
}
