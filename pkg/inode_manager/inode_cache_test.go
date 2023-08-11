package inode_manager

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/stretchr/testify/assert"
	"testing"
)

func testWithMount(t *testing.T, cb func(t *testing.T, ctx context.Context, inode *fs.Inode)) {
	tmpdir := t.TempDir()
	root := &fs.Inode{}
	opts := &fs.Options{}
	opts.OnAdd = func(ctx context.Context) {
		cb(t, ctx, root)
	}
	server, err := fs.Mount(tmpdir, root, opts)
	if err != nil {
		t.Fatalf("Cannot mount server. Error: %v", err)
	}
	_ = server.Unmount()
}

func TestInodeCache_GetOrInsert(t *testing.T) {
	testWithMount(t, func(t *testing.T, ctx context.Context, root *fs.Inode) {
		cache := InodeCache{}
		cache.Init(16)

		type args struct {
			key       string
			mode      uint32
			overwrite bool
		}
		testCases := []struct {
			name string
			args
			expected fs.StableAttr
		}{
			{
				"insert first",
				args{"a", fuse.S_IFDIR, false},
				fs.StableAttr{Mode: fuse.S_IFDIR, Ino: 16, Gen: 0},
			},
			{
				"repeat first",
				args{"a", fuse.S_IFREG, false},
				fs.StableAttr{Mode: fuse.S_IFDIR, Ino: 16, Gen: 0},
			},
			{
				"update gen",
				args{"a", fuse.S_IFLNK, true},
				fs.StableAttr{Mode: fuse.S_IFLNK, Ino: 16, Gen: 1},
			},
			{
				"insert another",
				args{"b", fuse.S_IFREG, false},
				fs.StableAttr{Mode: fuse.S_IFREG, Ino: 17, Gen: 0},
			},
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				builder := func() (fs.InodeEmbedder, error) {
					return &fs.Inode{}, nil
				}
				result, err := cache.GetOrInsert(ctx, tc.key, tc.mode, root, builder, tc.overwrite)
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result.StableAttr())
			})
		}
	})
}
