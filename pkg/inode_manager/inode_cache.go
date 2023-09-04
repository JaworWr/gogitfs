// Package inode_manager contains types meant for storage and lazy creation of Inodes.
package inode_manager

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	"sync"
)

// InodeCache combines AttrStore and InodeStore to create a cache managing both inodes and their attributes.
type InodeCache struct {
	lock       *sync.Mutex
	AttrStore  *AttrStore
	InodeStore *InodeStore
}

// Init performs initialization. initialIno is passed to AttrStore.Init.
func (m *InodeCache) Init(initialIno uint64) {
	m.lock = &sync.Mutex{}
	m.AttrStore = &AttrStore{}
	m.AttrStore.Init(initialIno)
	m.InodeStore = &InodeStore{}
	m.InodeStore.Init()
}

// GetOrInsert returns the Inode corresponding to the given key. If the key is absent or overwrite == true,
// a new node will be created with Attr generated from AttrStore and the specified file mode.
// Other parameters are simply passed to InodeStore.GetOrInsert.
// Note that setting overwrite == true will also set updateGen == true,
// which means that the created node will get a different generation number.
func (m *InodeCache) GetOrInsert(
	ctx context.Context,
	key string,
	mode uint32,
	parent fs.InodeEmbedder,
	builder func() (fs.InodeEmbedder, error),
	overwrite bool,
) (*fs.Inode, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	attr := m.AttrStore.GetOrInsert(key, overwrite)
	attr.Mode = mode
	node, err := m.InodeStore.GetOrInsert(ctx, key, attr, parent, builder, overwrite)
	return node, err
}
