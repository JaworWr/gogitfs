package inode_manager

import (
	"context"
	"fmt"
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
	hash fmt.Stringer,
	mode uint32,
	parent fs.InodeEmbedder,
	builder func() fs.InodeEmbedder,
	updateGen bool,
) *fs.Inode {
	attr := m.InoStore.GetOrInsert(hash, updateGen)
	attr.Mode = mode
	node := m.InodeStore.GetOrInsert(ctx, hash, attr, parent, builder)
	return node
}
