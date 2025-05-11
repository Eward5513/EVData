package common

import (
	"log"
	"math"
	"time"
)

const (
	RAW_DATA_DIR_PATH         = "D:/zhangteng2/raw_data"
	REFINED_RAW_DATA_DIR_PATH = "D:/zhangteng2/refined_raw_data"
	POINT_DATA_DIR_PATH       = "D:/zhangteng2/filtered_data1_UTC+8"

	TRACK_RAW_DATA_DIR_PATH     = "D:/zhangteng2/track_raw_data"
	TRACK_DATA_DIR_PATH         = "D:/zhangteng2/track_data"
	PARQUET_COUNT           int = 10017
)

var (
	TimeOffset    = time.Date(2021, 12, 21, 8, 0, 0, 0, time.UTC).UnixMilli() - 693244800000
	VehicleCount  int
	BATCH         int
	BATCH_SIZE    int
	MIN_LATITUDE  float64 = 30.6974858
	MAX_LATITUDE  float64 = 31.8610639
	MIN_LONGITUDE float64 = 120.8557546
	MAX_LONGITUDE float64 = 122.0158854
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
