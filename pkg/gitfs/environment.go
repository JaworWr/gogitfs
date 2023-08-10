package gitfs

import "gogitfs/pkg/inode_manager"

var commitCache *inode_manager.InodeCache
var branchCache *branchNodeCache
var initRun = false

var commitIno uint64 = 2 << 60
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
