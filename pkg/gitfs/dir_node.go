package gitfs

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"gogitfs/pkg/error_handler"
	"gogitfs/pkg/gitfs/internal/utils"
	"gogitfs/pkg/logging"
	"os"
	"syscall"
	"time"
)

const EntryTimeout = 10 * time.Minute

// dirNode represents a directory in the Git repository (a Git tree)
type dirNode struct {
	repoNode
	entry *object.TreeEntry
	tree  *object.Tree
	attr  fuse.Attr
}

func (n *dirNode) GetCallCtx() logging.CallCtx {
	info := utils.NodeCallCtx(n)
	if n.entry == nil {
		info["root"] = true
	} else {
		info["root"] = false
		info["mode"] = n.entry.Mode
	}
	return info
}

func newDirNode(repo *git.Repository, commit *object.Commit) (*dirNode, error) {
	node := &dirNode{}
	node.repo = repo

	tree, err := repo.TreeObject(commit.TreeHash)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve tree from commit %v: %w", commit.Hash, err)
	}
	node.tree = tree
	node.attr = utils.CommitAttr(commit)
	node.attr.Mode = 0555
	return node, nil
}

func newChildNode(parent *dirNode, childEntry *object.TreeEntry) (fs.InodeEmbedder, error) {
	mode, err := childEntry.Mode.ToOSFileMode()
	if err != nil {
		return nil, fmt.Errorf("cannot get file mode of %v: %w", childEntry.Name, err)
	}
	switch mode.Type() {
	case 0:
		file, err := parent.tree.TreeEntryFile(childEntry)
		if err != nil {
			return nil, fmt.Errorf("cannot get file %v: %w", childEntry.Name, err)
		}
		child := newFileNode(file, parent.attr)
		return child, nil
	case os.ModeDir:
		tree, err := parent.repo.TreeObject(childEntry.Hash)
		if err != nil {
			return nil, fmt.Errorf("cannot get subdirectory tree %v: %w", childEntry.Name, err)
		}
		child := &dirNode{}
		child.repo = parent.repo
		child.tree = tree
		child.entry = childEntry
		child.attr = parent.attr
		return child, nil
	case os.ModeSymlink:
		file, err := parent.tree.TreeEntryFile(childEntry)
		if err != nil {
			return nil, fmt.Errorf("cannot get file %v: %w", childEntry.Name, err)
		}
		target, err := file.Contents()
		if err != nil {
			return nil, fmt.Errorf("cannot read symlink %v: %w", childEntry.Name, err)
		}
		attr := parent.attr
		child := &fs.MemSymlink{Attr: attr, Data: []byte(target)}
		return child, nil
	default:
		logging.WarningLog.Printf("unsupported file type: %v (Git tree mode: %v)", mode.Type(), childEntry.Mode)
		return nil, nil
	}
}

// Getattr returns attributes corresponding to the given commit
func (n *dirNode) Getattr(_ context.Context, _ fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	logging.LogCall(n, nil)
	out.Attr = n.attr
	return fs.OK
}

type dirStream struct {
	// tree to iterate over
	tree *object.Tree
	// channel returning remaining entries
	rest <-chan *fuse.DirEntry
	// channel indicating that the iteration should stop
	stop chan<- int
}

// readTreeEntries reads the commits from `iter`, generates corresponding entries and places them in the channel `next`.
// If a value is read from `stop`, the function returns immediately.
func readTreeEntries(tree *object.Tree, next chan<- *fuse.DirEntry, stop <-chan int) {
	funcName := logging.CurrentFuncName(0, logging.Package)
	for _, treeEntry := range tree.Entries {
		logging.DebugLog.Printf(
			"%s: read entry %v",
			funcName,
			treeEntry.Name,
		)

		var entry fuse.DirEntry
		entry.Name = treeEntry.Name
		mode, err := treeEntry.Mode.ToOSFileMode()
		if err != nil {
			error_handler.Logging.HandleError(err)
		}
		entry.Mode = uint32(mode.Type())
		select {
		case <-stop:
			break
		case next <- &entry:
		}
	}
	close(next)
}

var _ fs.NodeGetattrer = (*dirNode)(nil)
