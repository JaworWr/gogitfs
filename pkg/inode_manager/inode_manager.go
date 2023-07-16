package inode_manager

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	"sync"
)

type InodeManager struct {
	lock       *sync.Mutex
	InoStore   *InoStore
	InodeStore *InodeStore
}

func (m *InodeManager) Init(initialIno uint64) {
	m.lock = &sync.Mutex{}
	m.InoStore = &InoStore{}
	m.InoStore.Init(initialIno)
	m.InodeStore = &InodeStore{}
	m.InodeStore.Init()
}

func (m *InodeManager) GetOrInsert(
	ctx context.Context,
	key string,
	mode uint32,
	parent fs.InodeEmbedder,
	builder func() (fs.InodeEmbedder, error),
	overwrite bool,
) (*fs.Inode, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	attr := m.InoStore.GetOrInsert(key, overwrite)
	attr.Mode = mode
	node, err := m.InodeStore.GetOrInsert(ctx, key, attr, parent, builder, overwrite)
	return node, err
}
