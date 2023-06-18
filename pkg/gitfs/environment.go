package gitfs

import "gogitfs/pkg/inode_manager"

var commitNodeMgr *inode_manager.InodeManager

func Init() {
	commitNodeMgr = inode_manager.InodeManager()
}
