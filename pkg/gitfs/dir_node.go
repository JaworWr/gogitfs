package gitfs

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/filemode"
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
	tree *object.Tree
	idx  int
}

// HasNext returns true if there are more entries.
func (s *dirStream) HasNext() bool {
	return s.idx < len(s.tree.Entries)
}

// Next returns the next entry.
func (s *dirStream) Next() (entry fuse.DirEntry, errno syscall.Errno) {
	if s.idx >= len(s.tree.Entries) {
		errno = syscall.ENOENT
		return
	}
	treeEntry := s.tree.Entries[s.idx]
	entry.Name = treeEntry.Name
	entry.Mode = convertMode(treeEntry.Mode)
	return
}

// Close closes the stream and cleans up any resources.
func (s *dirStream) Close() {

}

func convertMode(mode filemode.FileMode) uint32 {
	res, err := mode.ToOSFileMode()
	if err != nil {
		error_handler.Fatal.HandleError(fmt.Errorf("error converting tree entry mode: %w", err))
	}
	return uint32(res)
}

var _ fs.NodeGetattrer = (*dirNode)(nil)
