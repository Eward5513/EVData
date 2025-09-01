package CSV

import (
	"EVdata/common"
	"EVdata/proto_struct"
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"
)

func ReadRawPointFromCSV(path string, vin int) []*proto_struct.RawPoint {
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

	var points []*proto_struct.RawPoint
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("读取CSV记录失败:", err)
		}

		speed, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			log.Printf("解析speed失败[%s]: %v\n", record[1], err)
			continue
		}
		vehicleStatus, _ := strconv.Atoi(record[2])
		haveDriver, _ := strconv.Atoi(record[3])
		haveBrake, _ := strconv.Atoi(record[4])
		acceleratorPedal, _ := strconv.Atoi(record[5])
		brakeStatus, _ := strconv.Atoi(record[6])
		longitude, err := strconv.ParseFloat(record[8], 64)
		if err != nil {
			log.Printf("解析longitude失败[%s]: %v\n", record[2], err)
			continue
		}
		latitude, err := strconv.ParseFloat(record[9], 64)
		if err != nil {
			log.Printf("解析latitude失败[%s]: %v\n", record[3], err)
			continue
		}

		point := &proto_struct.RawPoint{
			Vin:              int32(vin),
			Time:             record[0],
			Speed:            speed,
			VehicleStatus:    int32(vehicleStatus),
			HaveDriver:       int32(haveDriver),
			HaveBrake:        int32(haveBrake),
			AcceleratorPedal: int32(acceleratorPedal),
			BrakeStatus:      int32(brakeStatus),
			Longitude:        longitude,
			Latitude:         latitude,
			TimeInt:          common.ParseTimeToInt(record[0]),
		}
		points = append(points, point)
	}

	//common.InfoLog("成功读取轨迹点数量: %d\n", len(points))
	return points
}

//
//func ReadTrackFromCSV(path string) (*common.Track, error) {
//	file, err := os.Open(path)
//	if err != nil {
//		return nil, fmt.Errorf("打开CSV文件失败: %v", err)
//	}
//	defer file.Close()
//
//	reader := csv.NewReader(file)
//	// 跳过表头
//	if _, err := reader.Read(); err != nil {
//		log.Println("error when reading csv:", err)
//	}
//
//	// 一个文件只有一条轨迹
//	var track *common.Track
//
//	for {
//		record, err := reader.Read()
//		if err == io.EOF {
//			break
//		}
//		if err != nil {
//			return nil, fmt.Errorf("读取CSV记录失败: %v", err)
//		}
//
//		// 解析轨迹段数据
//		vin, _ := strconv.Atoi(record[0])
//		tid, _ := strconv.Atoi(record[2])
//		roadID, _ := strconv.ParseInt(record[5], 10, 64)
//
//		var trackPoints []*proto_struct.RawPoint
//		var originalPoints []*proto_struct.RawPoint
//
//		// 解析JSON字符串为对应的结构体
//		if err := json.Unmarshal([]byte(record[6]), &trackPoints); err != nil {
//			log.Printf("解析轨迹点失败[%s]: %v\n", record[6], err)
//			continue
//		}
//		if err := json.Unmarshal([]byte(record[7]), &originalPoints); err != nil {
//			log.Printf("解析原始轨迹点失败[%s]: %v\n", record[7], err)
//			continue
//		}
//
//		// 创建轨迹段
//		seg := &common.TrackSegment{
//			StartTime:      record[3],
//			EndTime:        record[4],
//			RoadID:         roadID,
//			TrackPoints:    trackPoints,
//			OriginalPoints: originalPoints,
//		}
//
//		if track == nil {
//			track = &common.Track{
//				Vin:       vin,
//				Tid:       tid,
//				StartTime: record[2],
//				EndTime:   record[3],
//				Date:      record[1],
//				TrackSegs: make([]*common.TrackSegment, 0),
//			}
//		}
//		track.TrackSegs = append(track.TrackSegs, seg)
//
//		track.EndTime = record[3]
//	}
//
//	return track, nil
//}

func ReadMatchingPointFromCSV(path string) []*proto_struct.MatchingPoint {
	file, err := os.Open(path)
	if err != nil {
		common.ErrorLog("打开CSV文件失败:", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	if _, err := reader.Read(); err != nil {
		log.Println("error when reading csv:", err)
	}

	mps := make([]*proto_struct.MatchingPoint, 0)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			common.ErrorLog("读取CSV记录失败:", err)
		}

		mp := &proto_struct.MatchingPoint{}

		mp.OriginalLon, _ = strconv.ParseFloat(record[0], 64)
		mp.OriginalLat, _ = strconv.ParseFloat(record[1], 64)
		mp.MatchedLon, _ = strconv.ParseFloat(record[2], 64)
		mp.MatchedLat, _ = strconv.ParseFloat(record[3], 64)

		mps = append(mps, mp)
	}

	return mps
}

func ReadTrackPointFromCSV(path string) []*proto_struct.TrackPoint {
	file, err := os.Open(path)
	if err != nil {
		common.ErrorLog("打开CSV文件失败:", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	if _, err := reader.Read(); err != nil {
		log.Println("error when reading csv:", err)
	}

	tps := make([]*proto_struct.TrackPoint, 0)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			common.ErrorLog("读取CSV记录失败:", err)
		}

		tp := &proto_struct.TrackPoint{
			Time:    record[0],
			TimeInt: common.ParseTimeToInt(record[0]),
		}

		tp.Speed, _ = strconv.ParseFloat(record[1], 64)
		tp.Longitude, _ = strconv.ParseFloat(record[2], 64)
		tp.Latitude, _ = strconv.ParseFloat(record[3], 64)
		var temp int
		temp, _ = strconv.Atoi(record[4])
		tp.VehicleStatus = int32(temp)
		temp, _ = strconv.Atoi(record[5])
		tp.HaveDriver = int32(temp)
		temp, _ = strconv.Atoi(record[6])
		tp.HaveBrake = int32(temp)
		temp, _ = strconv.Atoi(record[7])
		tp.AcceleratorPedal = int32(temp)
		temp, _ = strconv.Atoi(record[8])
		tp.BrakeStatus = int32(temp)
		tp.MatchedLon, _ = strconv.ParseFloat(record[9], 64)
		tp.MatchedLat, _ = strconv.ParseFloat(record[10], 64)
		temp, _ = strconv.Atoi(record[11])
		tp.RoadId = int64(temp)
		temp, _ = strconv.Atoi(record[12])
		tp.IsBad = int32(temp)

		tps = append(tps, tp)
	}

	return tps
}
