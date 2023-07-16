package inode_manager

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	"sync"
)

type InodeStore struct {
	lock   *sync.Mutex
	inodes map[string]*fs.Inode
}

func (s *InodeStore) Init() {
	s.inodes = make(map[string]*fs.Inode)
	s.lock = &sync.Mutex{}
}

func (s *InodeStore) GetOrInsert(
	ctx context.Context,
	key string,
	attr fs.StableAttr,
	parent fs.InodeEmbedder,
	builder func() (fs.InodeEmbedder, error),
	overwrite bool,
) (*fs.Inode, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	inode, ok := s.inodes[key]
	if ok && !overwrite {
		return inode, nil
	}
	newEmb, err := builder()
	if err != nil {
		return nil, err
	}
	newNode := parent.EmbeddedInode().NewPersistentInode(ctx, newEmb, attr)
	s.inodes[key] = newNode
	return newNode, nil
}
