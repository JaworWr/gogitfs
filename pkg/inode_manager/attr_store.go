package inode_manager

import (
	"github.com/hanwen/go-fuse/v2/fs"
	"sync"
)

type attrEntry struct {
	ino, gen uint64
}

func (e *attrEntry) toStableAttr() fs.StableAttr {
	return fs.StableAttr{
		Ino: e.ino,
		Gen: e.gen,
	}
}

// AttrStore stores and generates inode and generation numbers for each key.
// Each new key gets the next available number.
type AttrStore struct {
	lock    *sync.Mutex
	nextIno uint64
	attrs   map[string]*attrEntry
}

// Init performs initialization. initialIno specifies the number that wil lbe received by the first added key.
func (s *AttrStore) Init(initialIno uint64) {
	s.lock = &sync.Mutex{}
	s.nextIno = initialIno
	s.attrs = make(map[string]*attrEntry)
}

// GetOrInsert for a given key returns fs.StableAttr containing inode and generation number for the given key.
// If the key was absent, it returns the next available inode number and generation number equal to 0.
// Subsequent calls will return the same inode number and the same generation number if updateGen == false.
// If updateGen == tre, the returned generation number will be increased.
func (s *AttrStore) GetOrInsert(key string, updateGen bool) fs.StableAttr {
	s.lock.Lock()
	defer s.lock.Unlock()
	attr, keyPresent := s.attrs[key]
	if keyPresent {
		if updateGen {
			attr.gen += 1
		}
	} else {
		attr = &attrEntry{ino: s.nextIno, gen: 0}
		s.nextIno += 1
		s.attrs[key] = attr
	}
	return attr.toStableAttr()
}
