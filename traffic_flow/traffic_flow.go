package traffic_flow

import (
	"EVdata/CSV"
	"EVdata/common"
	"EVdata/proto_struct"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	workerCount int = 500
)

func StartReader(rch chan []*proto_struct.TrackPoint) {

	for i := 1; i <= common.VEHICLE_COUNT; i++ {
		mps := CSV.ReadTrackPointFromCSV(i)
		rch <- mps
		if i%1000 == 0 {
			common.InfoLog("finish reading", i, "/", common.VEHICLE_COUNT)
		}
	}
	close(rch)
}

func StartWriter(wch chan []*common.TrafficFlow) {
	baseDir := common.TRACK_DATA_DIR_PATH

	// 打开三个输出文件
	queryFile, err := os.Create(filepath.Join(baseDir, "query.txt"))
	if err != nil {
		common.ErrorLog("Failed to create query.txt: " + err.Error())
		return
	}
	defer queryFile.Close()

	routeFile, err := os.Create(filepath.Join(baseDir, "route.txt"))
	if err != nil {
		common.ErrorLog("Failed to create route.txt: " + err.Error())
		return
	}
	defer routeFile.Close()

	timeFile, err := os.Create(filepath.Join(baseDir, "time.txt"))
	if err != nil {
		common.ErrorLog("Failed to create time.txt: " + err.Error())
		return
	}
	defer timeFile.Close()

	vinFile, err := os.Create(filepath.Join(baseDir, "vin.txt"))
	if err != nil {
		common.ErrorLog("Failed to create vin.txt: " + err.Error())
		return
	}
	defer vinFile.Close()

	// 处理传入的 TrafficFlow 数据
	for trafficFlows := range wch {
		for _, tf := range trafficFlows {
			if tf == nil {
				continue
			}

			// 写入 query.txt: vin 起点id 终点id 开始时间 时间戳字符串
			// 将 Unix 毫秒时间戳转换为时间字符串（UTC时区，与生成时保持一致）
			//timeStr := time.UnixMilli(tf.Time[0]).UTC().Format("15:04:05")
			//_, err := fmt.Fprintf(queryFile, "%d %d %d %d %s\n", tf.Vin, tf.Node[0], tf.Node[len(tf.Node)-1], tf.Time[0], timeStr)
			_, err := fmt.Fprintf(queryFile, "%d %d %d\n", tf.Node[0], tf.Node[len(tf.Node)-1], tf.Time[0])

			if err != nil {
				common.ErrorLog("Failed to write to query.txt: " + err.Error())
				continue
			}

			// 写入 route.txt: vin 节点数量 node1 bool1 bool2 bool3 bool4 node2 bool1 bool2 bool3 bool4 ...
			//fmt.Fprintf(routeFile, "%d %d", tf.Vin, len(tf.Node))
			fmt.Fprintf(routeFile, "%d", len(tf.Node))
			for _, nodeId := range tf.Node {
				// 每个node后跟四个bool变量，这里先设为默认值false false false false
				fmt.Fprintf(routeFile, " %d 0 0 0 0", nodeId)
			}
			fmt.Fprintf(routeFile, "\n")

			// 写入 time.txt: vin 节点数量-1 time[1] time[2] ...
			//fmt.Fprintf(timeFile, "%d %d", tf.Vin, len(tf.Node)-1)
			fmt.Fprintf(timeFile, "%d", len(tf.Node)-1)
			for i := 1; i < len(tf.Time); i++ {
				fmt.Fprintf(timeFile, " %d", tf.Time[i])
			}
			fmt.Fprintf(timeFile, "\n")

			fmt.Fprintln(vinFile, tf.Vin)
		}
	}
}

func StartWorker(rch chan []*proto_struct.TrackPoint, wch chan []*common.TrafficFlow) {
	wg := &sync.WaitGroup{}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go Worker(rch, wch, wg)
	}

	wg.Wait()
	close(wch)
}

func Worker(rch chan []*proto_struct.TrackPoint, wch chan []*common.TrafficFlow, wg *sync.WaitGroup) {

	defer wg.Done()

	for data := range rch {
		stps := SplitTrackPoint(data)
		//common.InfoLog("split track point", len(stps))

		var res []*common.TrafficFlow
		for _, stp := range stps {
			//common.InfoLog("split", i, len(stp), stp[0].Time, stp[len(stp)-1].Time)
			tf := &common.TrafficFlow{}
			for _, p := range stp {
				// start time
				if tf.Node == nil {
					//tf.Time = append(tf.Time, common.TimeStringToUnixMillis(p.Time))
					tf.Vin = p.Vin
				}

				if p.NodeId != 0 {
					//common.InfoLog("add graph node", p.NodeId)
					tf.Node = append(tf.Node, p.NodeId)
					tf.Time = append(tf.Time, common.TimeStringToUnixMillis(p.Time))
				}
			}
			if len(tf.Node) > 1 {
				res = append(res, tf)
			}
		}
		wch <- res
	}
}

func SplitTrackPoint(data []*proto_struct.TrackPoint) [][]*proto_struct.TrackPoint {
	res := make([][]*proto_struct.TrackPoint, 0)
	startIndex := -1
	var preTime int64 = -1
	for j := 0; j < len(data); j++ {
		if startIndex == -1 {
			if data[j].IsBad == 0 && data[j].VehicleStatus == 1 {
				startIndex = j
				preTime = data[j].TimeInt
			}
		} else if data[j].IsBad == 1 || data[j].TimeInt-preTime > 600 {
			temp := make([]*proto_struct.TrackPoint, 0, j-startIndex)
			temp = append(temp, data[startIndex:j]...)
			res = append(res, temp)
			startIndex = -1
			preTime = -1
		} else {
			preTime = data[j].TimeInt
		}
	}
	if startIndex != -1 {
		temp := make([]*proto_struct.TrackPoint, 0, len(data)-startIndex)
		temp = append(temp, data[startIndex:]...)
		res = append(res, temp)
	}
	return res
}

//func GenerateNetwork() {
//	// 构建道路图
//	graph := mapmatching.BuildGraph("shanghai_new.json")
//
//	// 统计节点数与边数（按方向统计）
//	nodeCount := len(graph)
//	edgeCount := 0
//	for _, gn := range graph {
//		edgeCount += len(gn.Next)
//	}
//
//	// 输出到 TRACK_DATA_DIR_PATH/network.txt
//	outPath := filepath.Join(common.TRACK_DATA_DIR_PATH, "network.txt")
//	f, err := os.Create(outPath)
//	if err != nil {
//		common.ErrorLog("Failed to create network.txt: " + err.Error())
//		return
//	}
//	defer f.Close()
//
//	w := bufio.NewWriter(f)
//	// 第一行：node数量 edge数量
//	fmt.Fprintf(w, "%d %d\n", nodeCount, edgeCount)
//
//	// 后续每行：nodeID1 nodeID2 roadID roadLength
//	for _, gn := range graph {
//		for _, road := range gn.Next {
//			length := common.Distance(gn.Lat, gn.Lon, road.Node.Lat, road.Node.Lon)
//			fmt.Fprintf(w, "%d %d %d %.6f\n", gn.Id, road.Node.Id, road.ID, length)
//		}
//	}
//	_ = w.Flush()
//}
