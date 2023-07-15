package inode_manager

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
)

type InodeManager struct {
	InoStore   *InoStore
	InodeStore *InodeStore
}

func (m *InodeManager) Init(initialIno uint64) {
	m.InoStore.Init(initialIno)
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
	attr := m.InoStore.GetOrInsert(key, overwrite)
	attr.Mode = mode
	node, err := m.InodeStore.GetOrInsert(ctx, key, attr, parent, builder, overwrite)
	return node, err
}
