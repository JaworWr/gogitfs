package inode_manager

import (
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAttrStore_GetOrInsert(t *testing.T) {
	store := AttrStore{}
	store.Init(16)

	type args struct {
		key       string
		updateGen bool
	}
	testCases := []struct {
		name string
		args
		expected fs.StableAttr
	}{
		{
			"insert first",
			args{"a", false},
			fs.StableAttr{Ino: 16, Gen: 0},
		},
		{
			"repeat first",
			args{"a", false},
			fs.StableAttr{Ino: 16, Gen: 0},
		},
		{
			"update gen",
			args{"a", true},
			fs.StableAttr{Ino: 16, Gen: 1},
		},
		{
			"insert another",
			args{"b", false},
			fs.StableAttr{Ino: 17, Gen: 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := store.GetOrInsert(tc.key, tc.updateGen)
			assert.Equal(t, tc.expected, result)
		})
	}
}
