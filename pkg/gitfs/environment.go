package gitfs

import "gogitfs/pkg/inode_manager"

// commitCache is an InodeCache storing all commit nodes. This allows us to avoid duplication of commitNode objects.
var commitCache *inode_manager.InodeCache

// branchCache is a branchNodeCache storing all branch nodes and updating them as needed.
var branchCache *branchNodeCache

// initRun tells us whether Init() has been called.
var initRun = false

// commitIno is the initial inode number for the commit nodes.
var commitIno uint64 = 2 << 60

// branchIno is the initial inode number for the branch nodes.
var branchIno uint64 = 2 << 59

func Init() {
	if initRun {
		return
	}
	commitCache = &inode_manager.InodeCache{}
	commitCache.Init(commitIno)
	branchCache = &branchNodeCache{}
	branchCache.init(branchIno)
	initRun = true
}
