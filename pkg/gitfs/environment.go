package gitfs

import "gogitfs/pkg/inode_manager"

var commitNodeMgr *inode_manager.InodeManager
var initRun = false

var commitIno uint64 = 2 << 60
var rootIno uint64 = 2 << 62

func Init() {
	if initRun {
		return
	}
	commitNodeMgr = inode_manager.NewInodeManager(commitIno)
	initRun = true
}
