package inode_manager

import (
	"context"
	"fmt"
	"github.com/hanwen/go-fuse/v2/fs"
)

type InodeManager struct {
	inoStore   *InoStore
	inodeStore *InodeStore
}

func NewInodeManager(initialIno uint64) *InodeManager {
	return &InodeManager{
		inoStore:   NewInoStore(initialIno),
		inodeStore: NewInodeStore(),
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
	attr := m.inoStore.GetOrInsert(hash, updateGen)
	attr.Mode = mode
	node := m.inodeStore.GetOrInsert(ctx, hash, attr, parent, builder)
	return node
}
