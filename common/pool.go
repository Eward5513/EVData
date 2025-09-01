package common

import (
	"EVdata/proto_struct"
	"sync"
)

type MemoryMap struct {
	m1 map[*proto_struct.Track]bool
	m2 map[*proto_struct.MatchingPoint]bool
	//m3 map[*proto_struct.RawPoint]bool
	m4 map[*CandidatePoint]bool
}

func NewMemoryMap() *MemoryMap {
	return &MemoryMap{
		m1: make(map[*proto_struct.Track]bool),
		m2: make(map[*proto_struct.MatchingPoint]bool),
		//m3: make(map[*proto_struct.RawPoint]bool),
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

	//for p := range m.m3 {
	//	rawPointPool.Put(p)
	//}
	for t := range m.m1 {
		t.Mps = make([]*proto_struct.MatchingPoint, 0)
		trackPool.Put(t)
	}

	m.m1 = make(map[*proto_struct.Track]bool)
	m.m2 = make(map[*proto_struct.MatchingPoint]bool)
	//m.m3 = make(map[*proto_struct.RawPoint]bool)
	m.m4 = make(map[*CandidatePoint]bool)
}

//func (m *MemoryMap) RecordRawPoints(ps []*proto_struct.RawPoint) {
//	for _, p := range ps {
//		m.m3[p] = true
//	}
//
//}

var trackPool = sync.Pool{
	New: func() interface{} {
		return &proto_struct.Track{
			Mps: make([]*proto_struct.MatchingPoint, 0),
			Rps: make([]*proto_struct.RawPoint, 0),
		}
	},
}

func (m *MemoryMap) GetTrack() *proto_struct.Track {
	t := GetTrack()
	m.m1[t] = true
	return t
}

func GetTrack() *proto_struct.Track {
	t := trackPool.Get().(*proto_struct.Track)
	t.Mps = make([]*proto_struct.MatchingPoint, 0)
	t.Tid = 0
	t.Probability = 0
	t.Vin = 0
	t.Rps = make([]*proto_struct.RawPoint, 0)
	return t
}

func PutTrack(t *proto_struct.Track) {
	for _, mp := range t.Mps {
		PutMatchingPoint(mp)
	}

	for _, rp := range t.Rps {
		PutRawPoint(rp)
	}
	t.Tid = 0
	t.Probability = 0
	t.Vin = 0
	t.Mps = make([]*proto_struct.MatchingPoint, 0)
	t.Rps = make([]*proto_struct.RawPoint, 0)
	trackPool.Put(t)
}

var rawPointPool = sync.Pool{
	New: func() interface{} {
		return &proto_struct.RawPoint{}
	},
}

func GetRawPoint() *proto_struct.RawPoint {
	t := rawPointPool.Get().(*proto_struct.RawPoint)

	t.Vin = 0
	t.Speed = 0
	t.Latitude = 0
	t.Longitude = 0
	t.VehicleStatus = 0
	t.HaveBrake = 0
	t.HaveBrake = 0
	t.AcceleratorPedal = 0
	t.BrakeStatus = 0

	return t
}

func PutRawPoint(rp *proto_struct.RawPoint) {
	rp.Vin = 0
	rp.Speed = 0
	rp.Latitude = 0
	rp.Longitude = 0
	rp.VehicleStatus = 0
	rp.HaveBrake = 0
	rp.HaveBrake = 0
	rp.AcceleratorPedal = 0
	rp.BrakeStatus = 0

	rawPointPool.Put(rp)
}

var candidatePointPool = sync.Pool{
	New: func() interface{} {
		return &CandidatePoint{}
	},
}

func (m *MemoryMap) GetCandidatePoint() *CandidatePoint {
	cp := candidatePointPool.Get().(*CandidatePoint)
	m.m4[cp] = true
	cp.Lat = 0
	cp.Lon = 0
	cp.OriginalPoint = nil
	cp.RoadID = 0
	cp.Ep = 0
	cp.Distance = 0
	cp.Vertex = nil
	cp.TT = 0
	return cp
}

var matchingPointPool = sync.Pool{
	New: func() interface{} {
		return &proto_struct.MatchingPoint{}
	},
}

func (m *MemoryMap) GetMatchingPoint() *proto_struct.MatchingPoint {
	p := GetMatchingPoint()
	m.m2[p] = true
	return p
}

func GetMatchingPoint() *proto_struct.MatchingPoint {
	p := matchingPointPool.Get().(*proto_struct.MatchingPoint)
	p.RoadId = 0
	p.OriginalLat = 0
	p.OriginalLon = 0
	p.MatchedLon = 0
	p.MatchedLat = 0
	return p
}

func PutMatchingPoint(p *proto_struct.MatchingPoint) {
	p.RoadId = 0
	p.OriginalLat = 0
	p.OriginalLon = 0
	p.MatchedLon = 0
	p.MatchedLat = 0
	matchingPointPool.Put(p)
}
