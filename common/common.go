package common

import (
	"log"
	"math"
	"time"
)

const (
	RAW_DATA_DIR_PATH          = "D:/data/dataset_origin_anting_hkust"
	REFINED_RAW_DATA_DIR_PATH  = "D:/zhangteng3/refined_raw_data"
	POINT_DATA_DIR_PATH        = "D:/zhangteng3/points/csv"
	RAW_POINT_PARQUET_PATH     = "D:/zhangteng3/points1/points.parquet"
	MATCHED_POINT_PARQUET_PATH = "D:/zhangteng3/matched_points/points.parquet"
	MATCHED_TRACK_PARQUET_PATH = "D:/zhangteng3/matched_points/tracks.bin"
	MATCHED_POINT_CSV_DIR      = "D:/zhangteng3/matched_points/csv"

	TRACK_RAW_DATA_DIR_PATH     = "D:/zhangteng2/track_raw_data"
	TRACK_DATA_DIR_PATH         = "D:/zhangteng2/track_data"
	PARQUET_COUNT           int = 10017
	SERVER_WORKER_COUNT         = 5000
	VEHICLE_COUNT               = 93469
)

var (
	MIN_LATITUDE  float64 = 100
	MAX_LATITUDE  float64 = 1
	MIN_LONGITUDE float64 = 100
	MAX_LONGITUDE float64 = 1
)

type Summary struct {
	cnt, minLocationState, maxLocationState, zeroLocationState, oneLocationState                            int32
	minCt, maxCt, zeroCt                                                                                    int64
	minLongitude, maxLongitude                                                                              float64
	minLatitude, maxLatitude                                                                                float64
	minSpeed, maxSpeed                                                                                      float64
	speedOver220Cnt, zeroLongitude, zeroLatitude, abnormalSpeed, invalidSpeed, nullSpeed, nullLocationState int
}

func NewSummary() *Summary {
	return &Summary{
		minCt:        math.MaxInt64,
		minLatitude:  math.MaxFloat64,
		minLongitude: math.MaxFloat64,
		minSpeed:     math.MaxFloat64,
	}
}

func (s *Summary) Operate(v *RawPoint) {
	s.cnt++
	if v.CollectionTime == 0 {
		s.zeroCt++
	} else {
		s.maxCt = max(s.maxCt, v.CollectionTime)
		s.minCt = min(s.minCt, v.CollectionTime)
	}
	switch {
	case v.Speed == 65534:
		s.abnormalSpeed++
	case v.Speed == 65535:
		s.invalidSpeed++
	case v.Speed >= 220:
		s.speedOver220Cnt++
	case v.Speed < 0:
		s.nullSpeed++
	default:
		s.maxSpeed = max(s.maxSpeed, v.Speed)
		s.minSpeed = min(s.minSpeed, v.Speed)
	}

	s.maxLocationState = max(s.maxLocationState, v.LocationState)
	s.minLocationState = min(s.minLocationState, v.LocationState)
	if v.LocationState == 0 {
		s.zeroLocationState++
	}
	if v.LocationState == 1 {
		s.oneLocationState++
	}
	if v.LocationState == -1 {
		s.nullLocationState++
	}

	if v.Longitude < 1 {
		s.zeroLongitude++
	} else {
		s.maxLongitude = max(s.maxLongitude, v.Longitude)
		s.minLongitude = min(s.minLongitude, v.Longitude)
	}

	if v.Latitude < 1 {
		s.zeroLatitude++
	} else {
		s.maxLatitude = max(s.maxLatitude, v.Latitude)
		s.minLatitude = min(s.minLatitude, v.Latitude)
	}
}

func (s *Summary) PrintData() {
	log.Println("Total rows in csv:", s.cnt)
	log.Printf("max collection_time:%d, timestamp : %s\n", s.maxCt, time.UnixMilli(s.maxCt).Format("2006-01-02 15:04:05"))
	log.Printf("min collection_time:%d, timestamp : %s\n", s.minCt, time.UnixMilli(s.minCt).Format("2006-01-02 15:04:05"))
	log.Println("collection_time is zero:", s.zeroCt)
	log.Printf("max speed:%f, min speed:%f, speed=65534:%d,speed=65535:%d speed>220: %d invalid speed :%d\n", s.maxSpeed, s.minSpeed, s.abnormalSpeed, s.invalidSpeed, s.speedOver220Cnt, s.nullSpeed)
	log.Printf("minLocationState:%d, maxLocationState:%d, zeroLocationState:%d, oneLocationState:%d, nullLocationState:%d\n", s.minLocationState, s.maxLocationState, s.zeroLocationState, s.oneLocationState, s.nullLocationState)
	log.Printf("max longitude: %f, min longitude:%f,longitude==0:%d\n", s.maxLongitude, s.minLongitude, s.zeroLongitude)
	log.Printf("maxLatitude:%f,minLatitude:%f,latitude==0:%d\n", s.maxLatitude, s.minLatitude, s.zeroLatitude)
}
