package CSV

import (
	"EVdata/common"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

func ReadPointFromCSV(path string) []*common.TrackPoint {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal("打开CSV文件失败:", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// 跳过表头
	if _, err := reader.Read(); err != nil {
		log.Println("error when reading csv:", err)
	}

	var points []*common.TrackPoint
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("读取CSV记录失败:", err)
		}

		vin, err := strconv.Atoi(record[0])
		if err != nil {
			log.Printf("解析vin失败[%s]: %v\n", record[0], err)
			continue
		}
		collectionTime, err := strconv.ParseInt(record[1], 10, 64)
		if err != nil {
			log.Printf("解析collection_time失败[%s]: %v\n", record[1], err)
			continue
		}
		hour, err := strconv.Atoi(record[4])
		if err != nil {
			log.Printf("解析hour失败[%s]: %v\n", record[4], err)
			continue
		}
		speed, err := strconv.ParseFloat(record[5], 64)
		if err != nil {
			log.Printf("解析speed失败[%s]: %v\n", record[5], err)
			continue
		}
		longitude, err := strconv.ParseFloat(record[6], 64)
		if err != nil {
			log.Printf("解析longitude失败[%s]: %v\n", record[6], err)
			continue
		}
		latitude, err := strconv.ParseFloat(record[7], 64)
		if err != nil {
			log.Printf("解析latitude失败[%s]: %v\n", record[7], err)
			continue
		}

		point := &common.TrackPoint{
			Vin:            vin,
			CollectionTime: collectionTime,
			Date:           record[2],
			Timestamp:      record[3],
			Hour:           hour,
			Speed:          speed,
			Longitude:      longitude,
			Latitude:       latitude,
		}
		points = append(points, point)
	}

	//common.InfoLog("成功读取轨迹点数量: %d\n", len(points))
	return points
}

func ReadTrackFromCSV(path string) (*common.Track, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开CSV文件失败: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// 跳过表头
	if _, err := reader.Read(); err != nil {
		log.Println("error when reading csv:", err)
	}

	// 一个文件只有一条轨迹
	var track *common.Track

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("读取CSV记录失败: %v", err)
		}

		// 解析轨迹段数据
		vin, _ := strconv.Atoi(record[0])
		tid, _ := strconv.Atoi(record[2])
		roadID, _ := strconv.ParseInt(record[5], 10, 64)

		var trackPoints []*common.TrackPoint
		var originalPoints []*common.TrackPoint

		// 解析JSON字符串为对应的结构体
		if err := json.Unmarshal([]byte(record[6]), &trackPoints); err != nil {
			log.Printf("解析轨迹点失败[%s]: %v\n", record[6], err)
			continue
		}
		if err := json.Unmarshal([]byte(record[7]), &originalPoints); err != nil {
			log.Printf("解析原始轨迹点失败[%s]: %v\n", record[7], err)
			continue
		}

		// 创建轨迹段
		seg := &common.TrackSegment{
			StartTime:      record[3],
			EndTime:        record[4],
			RoadID:         roadID,
			TrackPoints:    trackPoints,
			OriginalPoints: originalPoints,
		}

		if track == nil {
			track = &common.Track{
				Vin:       vin,
				Tid:       tid,
				StartTime: record[2],
				EndTime:   record[3],
				Date:      record[1],
				TrackSegs: make([]*common.TrackSegment, 0),
			}
		}
		track.TrackSegs = append(track.TrackSegs, seg)

		track.EndTime = record[3]
	}

	return track, nil
}
