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
type TrafficFlow struct {
	Vin  int32
	Node []int64
	Time []int64
}

func Distance(lat1, lon1, lat2, lon2 float64) float64 {
	//if AVX2Supported() {
	//	return DistanceAVX2(x1, y1, x2, y2)
	//}
	return Haversine(lat1, lon1, lat2, lon2)
}

//func P2lDistance(x1, y1, x2, y2, x3, y3 float64) float64 {
//	//if AVX2Supported() {
//	//	return P2lDistanceAVX2(x1, y1, x2, y2, x3, y3)
//	//}
//	return math.Abs((x2-x1)*(y3-y1)-(y2-y1)*(x3-x1)) / Distance(x1, y1, x2, y2)
//}

//func CalT(x1, y1, x2, y2, x3, y3 float64) float64 {
//	//if AVX2Supported() {
//	//	return CalTAVX2(x1, y1, x2, y2, x3, y3)
//	//}
//	return ((x2-x1)*(x3-x1) + (y2-y1)*(y3-y1)) / ((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1))
//}

//func CalP(x1, x2, y1, y2, tt float64) (float64, float64) {
//	//if AVX2Supported() {
//	//	return CalPAVX2(x1, y1, x2, y2, tt)
//	//}
//	return x1 + tt*(x2-x1), y1 + tt*(y2-y1)
//}

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

const EarthRadius = 6371008.8 // meters, WGS84 mean Earth radius

// Haversine distance (meters). Inputs are degrees.
func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
	phi1 := lat1 * math.Pi / 180.0
	phi2 := lat2 * math.Pi / 180.0
	dphi := (lat2 - lat1) * math.Pi / 180.0
	dlam := (lon2 - lon1) * math.Pi / 180.0

	sdphi := math.Sin(dphi / 2)
	sdlam := math.Sin(dlam / 2)

	a := sdphi*sdphi + math.Cos(phi1)*math.Cos(phi2)*sdlam*sdlam
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return EarthRadius * c
}

// metersPerDegree returns meters per 1 degree of latitude and longitude at a given latitude (radians).
func metersPerDegree(latRad float64) (mLat, mLon float64) {
	mLat = 111132.92 - 559.82*math.Cos(2*latRad) + 1.175*math.Cos(4*latRad) - 0.0023*math.Cos(6*latRad)
	mLon = 111412.84*math.Cos(latRad) - 93.5*math.Cos(3*latRad) + 0.118*math.Cos(5*latRad)
	return
}

// CanProjectOntoSegment computes the projection parameter t of A onto line BC
// using a local meters-per-degree planar mapping. It returns whether the
// perpendicular foot lies on the segment (t in [0,1] within eps), and the raw t.
// If B and C are identical, onSeg=false and t=0.
func CanProjectOntoSegment(
	latA, lonA, latB, lonB, latC, lonC float64,
) (onSeg bool, t float64) {
	const eps = 1e-9

	// reference latitude for scale
	phi0 := ((latB + latC) / 2.0) * math.Pi / 180.0
	mLat, mLon := metersPerDegree(phi0)

	// local planar coords (meters), using B as origin
	Bx, By := 0.0, 0.0
	Cx := (lonC - lonB) * mLon
	Cy := (latC - latB) * mLat
	Ax := (lonA - lonB) * mLon
	Ay := (latA - latB) * mLat

	vx, vy := Cx-Bx, Cy-By
	wx, wy := Ax-Bx, Ay-By

	vv := vx*vx + vy*vy
	if vv < eps {
		return false, 0.0 // degenerate segment
	}

	t = (wx*vx + wy*vy) / vv
	onSeg = (t >= -eps && t <= 1.0+eps)
	return
}

// FootPointAndDistance returns the clamped foot point on segment BC (lat/lon in degrees)
// and the great-circle distance (meters) from A to that foot.
// It accepts a t (e.g., from CanProjectOntoSegment) and clamps it into [0,1].
func FootPointAndDistance(
	latA, lonA, latB, lonB, latC, lonC float64, t float64,
) (footLat, footLon, distMeters float64) {
	// clamp t to the segment
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	// interpolate directly in degrees (equivalent to back-transform from local meters)
	footLat = latB + t*(latC-latB)
	footLon = lonB + t*(lonC-lonB)

	// distance A -> foot
	distMeters = Haversine(latA, lonA, footLat, footLon)
	return
}
