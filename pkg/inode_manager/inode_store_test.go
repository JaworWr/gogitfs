package inode_manager

import (
	"context"
	"errors"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInodeStore_GetOrInsert(t *testing.T) {
	testWithNode(t, func(t *testing.T, ctx context.Context, root *fs.Inode) {
		store := &InodeStore{}
		store.Init()

		type args struct {
			key       string
			overwrite bool
		}
		testCases := []struct {
			name string
			args
			shouldCreate bool
		}{
			{
				"insert first",
				args{"a", false},
				true,
			},
			{
				"repeat first",
				args{"a", false},
				false,
			},
			{
				"overwrite",
				args{"a", true},
				true,
			},
			{
				"insert another",
				args{"b", false},
				true,
			},
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				didCreate := false
				builder := func() (fs.InodeEmbedder, error) {
					didCreate = true
					return &fs.Inode{}, nil
				}
				_, err := store.GetOrInsert(ctx, tc.key, fs.StableAttr{}, root, builder, tc.overwrite)
				assert.NoError(t, err)
				assert.Equal(t, tc.shouldCreate, didCreate)
			})
		}
		t.Run("error", func(t *testing.T) {
			builder := func() (fs.InodeEmbedder, error) {
				return nil, errors.New("foo")
			}
			_, err := store.GetOrInsert(ctx, "zzz", fs.StableAttr{}, root, builder, false)
			assert.Error(t, err)
		})
	})
}
