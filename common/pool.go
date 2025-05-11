package common

import (
	"EVdata/proto_struct"
	"sync"
)

type MemoryMap struct {
	m1 map[*proto_struct.Track]bool
	m2 map[*proto_struct.TrackPoint]bool
	m3 map[*proto_struct.TrackSegment]bool
	m4 map[*CandidatePoint]bool
}

func NewMemoryMap() *MemoryMap {
	return &MemoryMap{
		m1: make(map[*proto_struct.Track]bool),
		m2: make(map[*proto_struct.TrackPoint]bool),
		m3: make(map[*proto_struct.TrackSegment]bool),
		m4: make(map[*CandidatePoint]bool),
	}
}

func (m *MemoryMap) RecordTrackPoint(p *proto_struct.TrackPoint) {
	m.m2[p] = true
}

func (m *MemoryMap) RecordCandidatePoint(p *CandidatePoint) {
	m.m4[p] = true
}

func (m *MemoryMap) Clear() {
	for tp := range m.m2 {
		trackPointPool.Put(tp)
	}
	for cpt := range m.m4 {
		candidatePointPool.Put(cpt)
	}
	for ts := range m.m3 {
		trackSegmentPool.Put(ts)
	}
	for t := range m.m1 {
		t.TrackSegs = t.TrackSegs[:0]

		trackPool.Put(t)
	}
}

var trackPointPool = sync.Pool{
	New: func() interface{} {
		return &proto_struct.TrackPoint{}
	},
}

func GetTrackPoint(mmp *MemoryMap) *proto_struct.TrackPoint {
	tp := trackPointPool.Get().(*proto_struct.TrackPoint)
	if mmp != nil {
		mmp.m2[tp] = true
	}
	return tp
}

func PutTrackPoint(p *proto_struct.TrackPoint) {
	trackPointPool.Put(p)
}

// 创建 proto_struct.TrackSegment 对象池
var trackSegmentPool = sync.Pool{
	New: func() interface{} {
		return &proto_struct.TrackSegment{
			TrackPoints:    make([]*proto_struct.TrackPoint, 0),
			OriginalPoints: make([]*proto_struct.TrackPoint, 0),
		}
	},
}

func GetTrackSegment(mmp *MemoryMap) *proto_struct.TrackSegment {
	ts := trackSegmentPool.Get().(*proto_struct.TrackSegment)
	if mmp != nil {
		mmp.m3[ts] = true
	}
	ts.TrackPoints = make([]*proto_struct.TrackPoint, 0)
	ts.OriginalPoints = make([]*proto_struct.TrackPoint, 0)
	return ts
}

func PutTrackSegment(p *proto_struct.TrackSegment) {
	for _, tp := range p.TrackPoints {
		PutTrackPoint(tp)
	}
	for _, op := range p.OriginalPoints {
		PutTrackPoint(op)
	}
	trackSegmentPool.Put(p)
}

var trackPool = sync.Pool{
	New: func() interface{} {
		return &proto_struct.Track{
			TrackSegs: make([]*proto_struct.TrackSegment, 0),
		}
	},
}

func GetTrack(mmp *MemoryMap) *proto_struct.Track {
	t := trackPool.Get().(*proto_struct.Track)
	if mmp != nil {
		mmp.m1[t] = true
	}
	return t
}

func PutTrack(t *proto_struct.Track) {
	for _, ts := range t.TrackSegs {
		PutTrackSegment(ts)
	}

	t.TrackSegs = t.TrackSegs[:0]

	trackPool.Put(t)
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
	return cp
}

func PutCandidatePoint(p *CandidatePoint) {
	candidatePointPool.Put(p)
}
