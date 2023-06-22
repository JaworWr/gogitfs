package inode_manager

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
)

type InodeManager struct {
	InoStore   *InoStore
	InodeStore *InodeStore
}

func NewInodeManager(initialIno uint64) *InodeManager {
	return &InodeManager{
		InoStore:   NewInoStore(initialIno),
		InodeStore: NewInodeStore(),
	}
}

func (m *InodeManager) GetOrInsert(
	ctx context.Context,
	key string,
	mode uint32,
	parent fs.InodeEmbedder,
	builder func() fs.InodeEmbedder,
	overwrite bool,
) *fs.Inode {
	attr := m.InoStore.GetOrInsert(key, overwrite)
	attr.Mode = mode
	node := m.InodeStore.GetOrInsert(ctx, key, attr, parent, builder, overwrite)
	return node
}
