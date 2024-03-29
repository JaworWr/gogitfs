package gitfs

import (
	"context"
	"errors"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_branchNodeCache_getOrInsert(t *testing.T) {
	Init()
	repo, extras := makeRepo(t)
	node := &repoNode{}
	node.repo = repo
	server, _ := mountNode(t, node, func(t *testing.T, ctx context.Context, inode *fs.Inode) {
		cache := &branchNodeCache{}
		cache.init(16)

		type args struct {
			branchName string
			branchHash plumbing.Hash
		}
		testCases := []struct {
			name string
			args
			expectedAttr fs.StableAttr
		}{
			{
				"insert foo",
				args{"foo_branch", extras.commits["foo"]},
				fs.StableAttr{Mode: fuse.S_IFDIR, Ino: 16, Gen: 0},
			},
			{
				"repeat foo",
				args{"foo_branch", extras.commits["foo"]},
				fs.StableAttr{Mode: fuse.S_IFDIR, Ino: 16, Gen: 0},
			},
			{
				"insert bar",
				args{"bar_branch", extras.commits["bar"]},
				fs.StableAttr{Mode: fuse.S_IFDIR, Ino: 17, Gen: 0},
			},
			{
				"update foo",
				args{"foo_branch", extras.commits["baz"]},
				fs.StableAttr{Mode: fuse.S_IFDIR, Ino: 16, Gen: 1},
			},
			{
				"update foo again",
				args{"foo_branch", extras.commits["foo"]},
				fs.StableAttr{Mode: fuse.S_IFDIR, Ino: 16, Gen: 2},
			},
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				reference := plumbing.NewHashReference(
					plumbing.NewBranchReferenceName(tc.branchName),
					tc.branchHash,
				)
				commit, inode, err := cache.getOrInsert(ctx, reference, node)
				assert.NoError(t, err, "unexpected error on running getOrInsert")
				assert.Equal(t, tc.branchHash, commit.Hash, "hashes are not equal")
				assert.Equal(t, tc.expectedAttr, inode.StableAttr(), "attributes are not equal")
			})
		}

		t.Run("nonexistent commit", func(t *testing.T) {
			reference := plumbing.NewHashReference(
				plumbing.NewBranchReferenceName("aaa"),
				plumbing.Hash{},
			)
			commit, inode, err := cache.getOrInsert(ctx, reference, node)
			assert.Error(t, err, "expected an error")
			assert.Nil(t, commit, "commit should be nil on error")
			assert.Nil(t, inode, "inode should be nil on error")
			assert.True(t, errors.Is(err, plumbing.ErrObjectNotFound), "error should be ErrObjectNotFound")
		})
	})
	_ = server.Unmount()
}
