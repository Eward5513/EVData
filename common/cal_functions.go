package common

import "golang.org/x/sys/cpu"

func AVX2Supported() bool {
	return cpu.X86.HasAVX2
}

//go:noescape
func DistanceAVX2(x1, y1, x2, y2 float64) float64

//go:noescape
func P2lDistanceAVX2(x1, y1, x2, y2, x3, y3 float64) float64

//go:noescape
func CalTAVX2(x1, y1, x2, y2, x3, y3 float64) float64

//go:noescape
func CalPAVX2(x1, x2, y1, y2, tt float64) (x, y float64)

//go:noescape
func CalEPAVX2(dis float64) float64
