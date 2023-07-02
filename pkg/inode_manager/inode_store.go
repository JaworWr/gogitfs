package inode_manager

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
)

type InodeStore struct {
	inodes map[string]*fs.Inode
}

func (s *InodeStore) Init() {
	s.inodes = make(map[string]*fs.Inode)
}

func (s *InodeStore) GetOrInsert(
	ctx context.Context,
	key string,
	attr fs.StableAttr,
	parent fs.InodeEmbedder,
	builder func() fs.InodeEmbedder,
	overwrite bool,
) *fs.Inode {
	inode, ok := s.inodes[key]
	if ok && !overwrite {
		return inode
	}
	newEmb := builder()
	newNode := parent.EmbeddedInode().NewPersistentInode(ctx, newEmb, attr)
	s.inodes[key] = newNode
	return newNode
}
