package inode_manager

import (
	"github.com/hanwen/go-fuse/v2/fs"
)

type InoStore struct {
	nextIno uint64
	inos    map[string]uint64
	gens    map[string]uint64
}

func (s *InoStore) Init(initialIno uint64) {
	s.nextIno = initialIno
	s.inos = make(map[string]uint64)
	s.gens = make(map[string]uint64)
}

func (s *InoStore) GetOrInsert(key string, updateGen bool) fs.StableAttr {
	ino, ok := s.inos[key]
	if ok {
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
