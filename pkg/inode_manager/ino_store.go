package inode_manager

import (
	"github.com/hanwen/go-fuse/v2/fs"
	"sync"
)

// InoStore stores and generates inode and generation numbers for each key.
// Each new key gets the next available number.
type InoStore struct {
	lock    *sync.Mutex
	nextIno uint64
	inos    map[string]uint64
	gens    map[string]uint64
}

// Init performs initialization. initialIno specifies the number that wil lbe received by the first added key.
func (s *InoStore) Init(initialIno uint64) {
	s.lock = &sync.Mutex{}
	s.nextIno = initialIno
	s.inos = make(map[string]uint64)
	s.gens = make(map[string]uint64)
}

// GetOrInsert for a given key returns fs.StableAttr containing inode and generation number for the given key.
// If the key was absent, it returns the next available inode number and generation number equal to 0.
// Subsequent calls will return the same inode number and the same generation number if updateGen == false.
// If updateGen == tre, the returned generation number will be increased.
func (s *InoStore) GetOrInsert(key string, updateGen bool) fs.StableAttr {
	s.lock.Lock()
	defer s.lock.Unlock()
	ino, keyPresent := s.inos[key]
	if keyPresent {
		gen := s.gens[key]
		if updateGen {
			gen += 1
			s.gens[key] += 1
		}
		return fs.StableAttr{Ino: ino, Gen: gen}
	}
	result := fs.StableAttr{Ino: s.nextIno, Gen: 0}
	s.inos[key] = s.nextIno
	s.nextIno += 1
	s.gens[key] = 0
	return result
}
