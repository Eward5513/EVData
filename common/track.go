package common

import (
	"EVdata/proto_struct"
	"math"
)

const (
	MAGIC_NUM = 111320
	SIGMA     = 40
)

type CandidateType = int

const (
	NORMAL   CandidateType = iota
	DISCRETE CandidateType = iota
)

type CandidatePoint struct {
	Vertex        []*GraphNode
	TT            float64
	Ttype         CandidateType
	Distance      float64
	Lat, Lon      float64
	Ep            float64
	RoadID        int64
	OriginalPoint *proto_struct.RawPoint
}

type CandidateSet struct {
	Cp   []*CandidatePoint
	Next *CandidateSet
}

type Element struct {
	Type      string  `json:"type"`
	ID        int64   `json:"id"`
	Timestamp string  `json:"timestamp"`
	Version   int     `json:"version"`
	Changeset int64   `json:"changeset"`
	User      string  `json:"user"`
	UID       int     `json:"uid"`
	Nodes     []int64 `json:"nodes,omitempty"`
	Tags      Tags    `json:"tags,omitempty"`
	Lat       float64 `json:"Lat,omitempty"`
	Lon       float64 `json:"Lon,omitempty"`
}

type Tags struct {
	Highway  string `json:"highway,omitempty"`
	Name     string `json:"name,omitempty"`
	NameZh   string `json:"name:zh,omitempty"`
	Surface  string `json:"surface,omitempty"`
	Lanes    string `json:"lanes,omitempty"`
	MaxSpeed string `json:"maxspeed,omitempty"`
	Oneway   string `json:"oneway,omitempty"`
}

type Path struct {
	Points     []*GraphNode
	StartPoint *GraphNode
	EndPoint   *GraphNode
	Distance   float64
}

type GraphNode struct {
	Lat, Lon     float64
	Next         []*Road
	Id           int64
	ShortestPath map[*GraphNode]*Path
}

type Road struct {
	ID   int64
	Way  *Element
	Node *GraphNode
}

func Distance(x1, y1, x2, y2 float64) float64 {
	//if AVX2Supported() {
	//	return DistanceAVX2(x1, y1, x2, y2)
	//}
	return math.Hypot(x2-x1, y2-y1)
}

func P2lDistance(x1, y1, x2, y2, x3, y3 float64) float64 {
	//if AVX2Supported() {
	//	return P2lDistanceAVX2(x1, y1, x2, y2, x3, y3)
	//}
	return math.Abs((x2-x1)*(y3-y1)-(y2-y1)*(x3-x1)) / Distance(x1, y1, x2, y2)
}

func CalT(x1, y1, x2, y2, x3, y3 float64) float64 {
	//if AVX2Supported() {
	//	return CalTAVX2(x1, y1, x2, y2, x3, y3)
	//}
	return ((x2-x1)*(x3-x1) + (y2-y1)*(y3-y1)) / ((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1))
}

func CalP(x1, x2, y1, y2, tt float64) (float64, float64) {
	//if AVX2Supported() {
	//	return CalPAVX2(x1, y1, x2, y2, tt)
	//}
	return x1 + tt*(x2-x1), y1 + tt*(y2-y1)
}

func CalEP(dis float64) float64 {
	//if AVX2Supported() {
	//	return math.Exp(CalEPAVX2(dis))
	//}
	//math.Exp(-(dis * dis) / (2 * SIGMA * SIGMA) / (SIGMA * math.Sqrt(2*math.Pi)))
	//const x = -3.1167365656361935e-06
	return math.Exp(-(dis * dis) / (2 * SIGMA * SIGMA) / (SIGMA * math.Sqrt(2*math.Pi)))
}

func Abs[T ~int64 | int](a, b T) T {
	if a < b {
		return b - a
	}
	return a - b
}
