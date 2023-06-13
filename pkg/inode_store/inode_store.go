package inode_store

import (
	"context"
	"fmt"
	"github.com/hanwen/go-fuse/v2/fs"
)

type InodeStore struct {
	inodes map[string]*fs.Inode
}

func NewInodeStore() *InodeStore {
	store := InodeStore{
		inodes: make(map[string]*fs.Inode),
	}
	return &store
}

func (s *InodeStore) GetOrInsert(
	ctx context.Context,
	hash fmt.Stringer,
	attr fs.StableAttr,
	parent fs.InodeEmbedder,
	builder func() fs.InodeEmbedder,
) *fs.Inode {
	hashStr := hash.String()
	inode, ok := s.inodes[hashStr]
	if ok {
		return inode
	}
	newEmb := builder()
	newNode := parent.EmbeddedInode().NewPersistentInode(ctx, newEmb, attr)
	s.inodes[hashStr] = newNode
	return newNode
}
