package gitfs

import "gogitfs/pkg/inode_manager"

var commitNodeMgr *inode_manager.InodeManager
var branchNodeMgr *branchNodeManager
var initRun = false

var commitIno uint64 = 2 << 60
var branchIno uint64 = 2 << 59

func Init() {
	if initRun {
		return
	}
	commitNodeMgr = &inode_manager.InodeManager{}
	commitNodeMgr.Init(commitIno)
	branchNodeMgr = newBranchNodeManager(branchIno)
	initRun = true
}
