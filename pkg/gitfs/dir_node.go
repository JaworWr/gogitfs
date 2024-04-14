package gitfs

import (
	"context"
	"errors"
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

const TreeEntryTimeout = 10 * time.Minute

var UnsupportedTreeNodeType = errors.New("unsupported tree node type")

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

func newTreeDirNode(repo *git.Repository, commit *object.Commit) (*dirNode, error) {
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

func newTreeChildNode(parent *dirNode, childEntry *object.TreeEntry) (fs.InodeEmbedder, fuse.Attr, error) {
	mode, err := childEntry.Mode.ToOSFileMode()
	if err != nil {
		return nil, fuse.Attr{}, fmt.Errorf("cannot get file mode of %v: %w", childEntry.Name, err)
	}
	attr := parent.attr
	switch mode.Type() {
	case 0:
		file, err := parent.tree.TreeEntryFile(childEntry)
		if err != nil {
			return nil, fuse.Attr{}, fmt.Errorf("cannot get file %v: %w", childEntry.Name, err)
		}
		child := newTreeFileNode(file, parent.attr)
		attr.Mode = fuse.S_IFREG | 0444
		return child, attr, nil
	case os.ModeDir:
		tree, err := parent.repo.TreeObject(childEntry.Hash)
		if err != nil {
			return nil, fuse.Attr{}, fmt.Errorf("cannot get subdirectory tree %v: %w", childEntry.Name, err)
		}
		child := &dirNode{}
		child.repo = parent.repo
		child.tree = tree
		child.entry = childEntry
		child.attr = parent.attr
		return child, attr, nil
	case os.ModeSymlink:
		file, err := parent.tree.TreeEntryFile(childEntry)
		if err != nil {
			return nil, fuse.Attr{}, fmt.Errorf("cannot get file %v: %w", childEntry.Name, err)
		}
		target, err := file.Contents()
		if err != nil {
			return nil, fuse.Attr{}, fmt.Errorf("cannot read symlink %v: %w", childEntry.Name, err)
		}
		attr := parent.attr
		child := &fs.MemSymlink{Attr: attr, Data: []byte(target)}
		attr.Mode = fuse.S_IFLNK
		return child, attr, nil
	default:
		return nil, fuse.Attr{}, fmt.Errorf("%w: %v", UnsupportedTreeNodeType, mode)
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

// Readdir returns the contents of the directory.
func (n *dirNode) Readdir(_ context.Context) (fs.DirStream, syscall.Errno) {
	s := &dirStream{n.tree, 0}
	return s, fs.OK
}

func (n *dirNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	entry, err := n.tree.FindEntry(name)
	if err != nil {
		if errors.Is(err, object.ErrEntryNotFound) {
			logging.WarningLog.Printf("Name %v not found", name)
			return nil, syscall.ENOENT
		} else {
			error_handler.Logging.HandleError(fmt.Errorf("cannot get tree entry %v: %w", name, err))
			return nil, syscall.EIO
		}
	}

	node, attr, err := newTreeChildNode(n, entry)
	if err != nil {
		error_handler.Logging.HandleError(fmt.Errorf("cannot create tree node: %w", err))
		return nil, syscall.EIO
	}

	out.SetEntryTimeout(TreeEntryTimeout)
	out.Attr = attr

	child := n.NewInode(ctx, node, fs.StableAttr{Mode: out.Mode & syscall.S_IFMT})
	return child, fs.OK
}

var _ fs.NodeGetattrer = (*dirNode)(nil)
var _ fs.NodeReaddirer = (*dirNode)(nil)
var _ fs.NodeLookuper = (*dirNode)(nil)
