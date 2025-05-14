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
		ts.OriginalPoints = make([]*proto_struct.TrackPoint, 0)
		ts.TrackPoints = make([]*proto_struct.TrackPoint, 0)
		trackSegmentPool.Put(ts)
	}
	for t := range m.m1 {
		t.TrackSegs = make([]*proto_struct.TrackSegment, 0)
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
	tp.Latitude = 0
	tp.Longitude = 0
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
	t.TrackSegs = make([]*proto_struct.TrackSegment, 0)
	t.IsBad = 0
	t.DisCount = 0
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
