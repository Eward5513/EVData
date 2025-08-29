package common

import (
	"EVdata/proto_struct"
	"sync"
)

type MemoryMap struct {
	m1 map[*proto_struct.Track]bool
	m2 map[*proto_struct.MatchingPoint]bool
	m4 map[*CandidatePoint]bool
}

func NewMemoryMap() *MemoryMap {
	return &MemoryMap{
		m1: make(map[*proto_struct.Track]bool),
		m2: make(map[*proto_struct.MatchingPoint]bool),
		m4: make(map[*CandidatePoint]bool),
	}
}

func (m *MemoryMap) Clear() {
	for cpt := range m.m4 {
		candidatePointPool.Put(cpt)
	}

	for p := range m.m2 {
		matchingPointPool.Put(p)
	}
	for t := range m.m1 {
		t.Tps = make([]*proto_struct.MatchingPoint, 0)
		trackPool.Put(t)
	}
}

var trackPool = sync.Pool{
	New: func() interface{} {
		return &proto_struct.Track{
			Tps: make([]*proto_struct.MatchingPoint, 0),
		}
	},
}

func GetTrack(mmp *MemoryMap) *proto_struct.Track {
	t := trackPool.Get().(*proto_struct.Track)
	if mmp != nil {
		mmp.m1[t] = true
	}
	t.Tps = make([]*proto_struct.MatchingPoint, 0)
	return t
}

var candidatePointPool = sync.Pool{
	New: func() interface{} {
		return &CandidatePoint{}
	},
}

func GetCandidatePoint(mmp *MemoryMap) *CandidatePoint {
	cp := candidatePointPool.Get().(*CandidatePoint)
	if mmp != nil {
		mmp.m4[cp] = true
	}
	cp.Lat = 0
	cp.Lon = 0
	return cp
}

var matchingPointPool = sync.Pool{
	New: func() interface{} {
		return &proto_struct.MatchingPoint{}
	},
}

func GetMatchingPoint(mmp *MemoryMap) *proto_struct.MatchingPoint {
	p := matchingPointPool.Get().(*proto_struct.MatchingPoint)
	if mmp != nil {
		mmp.m2[p] = true
	}
	p.RoadId = 0
	p.OriginalLat = 0
	p.OriginalLon = 0
	p.MatchedLon = 0
	p.MatchedLat = 0
	return p
}
