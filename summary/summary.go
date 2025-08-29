package main

import (
	"EVdata/common"
	"EVdata/pgpg"
	"log"
	"math"
	"time"
)

func main() {
	common.CreatLogFile("summary.log")
	defer common.CloseLogFile()

	start := time.Now()
	SummarizePoint()
	common.InfoLog("运行时长:", time.Since(start))

}

func SummarizePoint() {
	ps := pgpg.ReadPointFromParquet("D:/zhangteng3/points1/points.parquet")
	common.InfoLog("Total points: ", len(ps))
	var minLat, minLon, maxLat, maxLon float64
	minLat, minLon = math.MaxFloat64, math.MaxFloat64
	for _, p := range ps {
		minLat = min(minLat, p.Latitude)
		minLon = min(minLon, p.Longitude)
		maxLat = max(maxLat, p.Latitude)
		maxLon = max(maxLon, p.Longitude)
	}
	log.Println(minLat, minLon, maxLat, maxLon)
}
