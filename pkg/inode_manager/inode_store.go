package inode_manager

import (
	"context"
	"fmt"
	"github.com/hanwen/go-fuse/v2/fs"
	"sync"
)

// InodeStore stores a unique Inode per key.
type InodeStore struct {
	lock   *sync.Mutex
	inodes map[string]*fs.Inode
}

func (s *InodeStore) Init() {
	s.inodes = make(map[string]*fs.Inode)
	s.lock = &sync.Mutex{}
}

// GetOrInsert returns the Inode corresponding to the given key, or creates one if it doesn't exist.
// Subsequent calls will return the same node, unless called with overwrite == true.
// When the key is absent or overwrite == true, new node will be created by calling builder()
// and then calling NewPersistentInode on parent passing attr and the result.
// builder() will not be called if key is present and overwrite == false. Use overwrite to force creation
// of a new node.
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
		return nil, fmt.Errorf("cannot build an INode: %w", err)
	}
	newNode := parent.EmbeddedInode().NewPersistentInode(ctx, newEmb, attr)
	s.inodes[key] = newNode
	return newNode, nil
}
