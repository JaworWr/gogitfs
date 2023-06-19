package gitfs

import "gogitfs/pkg/inode_manager"

var commitNodeMgr *inode_manager.InodeManager
var initRun = false

func Init() {
	if initRun {
		return
	}
	commitNodeMgr = inode_manager.NewInodeManager(2 << 60)
	initRun = true
}
