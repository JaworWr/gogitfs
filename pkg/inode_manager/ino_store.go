package inode_manager

import (
	"fmt"
	"github.com/hanwen/go-fuse/v2/fs"
)

type InoStore struct {
	nextIno uint64
	inos    map[string]uint64
	gens    map[string]uint64
}

func NewInoStore(initialIno uint64) *InoStore {
	return &InoStore{
		nextIno: initialIno,
		inos:    make(map[string]uint64),
		gens:    make(map[string]uint64),
	}
}

func (s *InoStore) GetOrInsert(hash fmt.Stringer, updateGen bool) fs.StableAttr {
	hashStr := hash.String()
	ino, ok := s.inos[hashStr]
	if ok {
		gen := s.gens[hashStr]
		if updateGen {
			gen += 1
			s.gens[hashStr] += 1
		}
		return fs.StableAttr{Ino: ino, Gen: gen}
	}
	result := fs.StableAttr{Ino: s.nextIno, Gen: 0}
	s.inos[hashStr] = s.nextIno
	s.nextIno += 1
	s.gens[hashStr] = 0
	return result
}
