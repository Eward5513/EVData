package main

import (
	"EVdata/common"
	"EVdata/pgpg"
	"EVdata/proto_struct"
	"EVdata/proto_tools"
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
	"time"
)

// 道路网络数据结构
type RoadNetwork struct {
	Version   float64          `json:"version"`
	Generator string           `json:"generator"`
	Osm3s     Osm3s            `json:"osm3s"`
	Elements  []common.Element `json:"elements"`
}

type Osm3s struct {
	TimestampOsmBase string `json:"timestamp_osm_base"`
	Copyright        string `json:"copyright"`
}

type IndexNode struct {
	next       map[int64]*IndexNode
	graphNodes []*common.GraphNode
}

var (
	workerCount       int
	gcFrequency       int
	readerChannelSize int
	writerChannelSize int
	DistanceOffset    float64
)

func main() {

	begin := time.Now()

	var profMode bool
	var gcPercent int
	flag.BoolVar(&profMode, "p", false, "pprof mode")
	flag.Float64Var(&DistanceOffset, "d", 40, "Distance offset")
	flag.IntVar(&workerCount, "wc", 500, "worker count")
	flag.IntVar(&gcPercent, "gc", 100, "SetGCPercent")
	flag.IntVar(&gcFrequency, "gcf", -1, "SetGCFrequency")
	flag.IntVar(&common.BATCH_SIZE, "bs", 100, "batch size")
	flag.IntVar(&common.BATCH, "bi", -1, "batch id")
	flag.IntVar(&readerChannelSize, "rz", 1e4, "reader channel size")
	flag.IntVar(&writerChannelSize, "wz", 1e2, "writer channel size")
	flag.Parse()

	common.CreatLogFile("mapmatching.log")
	defer common.CloseLogFile()

	if profMode {
		//common.SetLogLevel(common.DEBUG)
		go func() {
			log.Println("start pprof tool")
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	//GC
	debug.SetGCPercent(gcPercent)

	//if err := common.CreateDirs(common.TRACK_RAW_DATA_DIR_PATH); err != nil {
	//	log.Fatal("error creating dirs:" + err.Error())
	//}

	//if err := os.Mkdir(common.TRACK_RAW_DATA_DIR_PATH, os.ModeDir); err != nil {
	//	common.ErrorLog("Error when creating target dir:" + err.Error())
	//}

	common.SetLogLevel(common.INFO)

	t := time.Now()
	graph := BuildGraph("roads_data.json")
	common.InfoLog("time for building graph: ", time.Since(t))
	t = time.Now()

	indexRoot := BuildIndex(graph)
	common.InfoLog("time for building index: ", time.Since(t))
	t = time.Now()

	PreComputing(graph)
	common.InfoLog("time for precomputing: ", time.Since(t))

	if common.BATCH != -1 {
		common.InfoLog("process batch data", common.BATCH)

		common.CloseLogFile()
		common.CreatLogFile("mapmatching" + strconv.Itoa(common.BATCH) + ".log")
		ProcessDataLoop(indexRoot)

	} else {
		//for i := 0; i <= common.PARQUET_COUNT/common.BATCH_SIZE; i++ {
		//	common.InfoLog("start loop: ", i)
		//	common.BATCH = i
		//	loopStartTime := time.Now()
		//
		//	ProcessDataLoop(indexRoot, readerChannelSize)
		//
		//	debug.FreeOSMemory()
		//	common.InfoLog("end loop: ", time.Since(loopStartTime))
		//	time.Sleep(time.Second * 20)
		//}

		common.BATCH = 0
		common.BATCH_SIZE = common.PARQUET_COUNT + 1
		if err := os.RemoveAll(common.TRACK_RAW_DATA_DIR_PATH); err != nil {
			common.FatalLog("Error removing target directory", err.Error())
		}
		if err := os.Mkdir(common.TRACK_RAW_DATA_DIR_PATH, os.ModeDir); err != nil {
			common.FatalLog("Error when creating target dir:" + err.Error())
		}
		ProcessDataLoop(indexRoot)
	}

	//if err := common.DeleteEmptyDirs(common.TRACK_DATA_DIR_PATH); err != nil {
	//	log.Fatal("error deleting empty dirs:" + err.Error())
	//}

	log.Println("Total Execution time: ", time.Now().Sub(begin))
}

func ProcessDataLoop(indexRoot *IndexNode) {
	wg := sync.WaitGroup{}
	readerChan := make(chan []*proto_struct.TrackPoint, readerChannelSize)
	writerChan := make(chan *proto_struct.Track, writerChannelSize)

	wg.Add(1)
	go StartReader(readerChan, &wg)
	wg.Add(1)
	go StartWriterManager(writerChan, &wg)

	ProcessData(readerChan, writerChan, indexRoot)

	wg.Wait()
}

func BuildGraph(jsonFile string) map[int64]*common.GraphNode {
	f, err := os.Open(jsonFile)
	defer f.Close()
	if err != nil {
		log.Fatal("打开json文件失败:", err)
	}
	byteValue, err := io.ReadAll(f)
	if err != nil {
		log.Fatal("读取json文件失败:", err)
	}

	var roadNetwork RoadNetwork
	if err := json.Unmarshal(byteValue, &roadNetwork); err != nil {
		log.Fatal("解析json数据失败:", err)
	}

	roads := make(map[int64]*common.Element)
	nodes := make(map[int64]*common.Element)
	graphNodes := make(map[int64]*common.GraphNode)

	for _, element := range roadNetwork.Elements {
		if element.Type == "way" {
			roads[element.ID] = &element
		} else if element.Type == "node" {
			nodes[element.ID] = &element
			graphNodes[element.ID] = &common.GraphNode{
				Lat:          element.Lat,
				Lon:          element.Lon,
				Next:         make([]*common.Road, 0),
				Id:           element.ID,
				ShortestPath: make(map[*common.GraphNode]*common.Path),
			}
		}
	}

	for _, road := range roads {
		for i := 0; i < len(road.Nodes)-1; i++ {
			if road.Tags.Oneway == "yes" {
				graphNodes[road.Nodes[i]].Next = append(graphNodes[road.Nodes[i]].Next, &common.Road{ID: road.ID, Way: road, Node: graphNodes[road.Nodes[i+1]]})
			} else {
				graphNodes[road.Nodes[i]].Next = append(graphNodes[road.Nodes[i]].Next, &common.Road{ID: road.ID, Way: road, Node: graphNodes[road.Nodes[i+1]]})
				graphNodes[road.Nodes[i+1]].Next = append(graphNodes[road.Nodes[i+1]].Next, &common.Road{ID: road.ID, Way: road, Node: graphNodes[road.Nodes[i]]})
			}
		}
	}
	log.Printf("版本: %v\n", roadNetwork.Version)
	log.Printf("生成器: %s\n", roadNetwork.Generator)
	log.Printf("元素数量: %d\n", len(roadNetwork.Elements))
	return graphNodes
}

func BuildIndex(graph map[int64]*common.GraphNode) *IndexNode {
	root := &IndexNode{}
	var buildIndex func(ro *IndexNode, minLat, maxLat, minLon, maxLon, base int64)
	buildIndex = func(ro *IndexNode, minLat, maxLat, minLon, maxLon, base int64) {
		//小数点后两位
		if base > 100000 {
			return
		}
		ro.next = make(map[int64]*IndexNode)
		var x, y int64
		for x = minLat; x <= maxLat; x++ {
			for y = minLon; y <= maxLon; y++ {
				var ne IndexNode
				ro.next[x*base+y] = &ne
				buildIndex(&ne, x*10, (x+1)*10-1, y*10, (y+1)*10-1, base*10)
			}
		}
	}
	buildIndex(root, int64(common.MIN_LATITUDE), int64(common.MAX_LATITUDE), int64(common.MIN_LONGITUDE), int64(common.MAX_LONGITUDE), 1000)

	for _, n := range graph {
		//if n.id == 330788984 {
		//	//31.1589551 121.6582715
		//	common.DebugLog("search node", n.lat, n.lon)
		//}
		indexNode := SearchIndex(n.Lat, n.Lon, root)
		//if n.id == 330788984 {
		//	common.DebugLog("index node", len(indexNode.graphNodes))
		//}
		indexNode.graphNodes = append(indexNode.graphNodes, n)
		//if n.id == 330788984 {
		//	common.DebugLog("append node", len(indexNode.graphNodes), indexNode.graphNodes[0].lon, indexNode.graphNodes[0].lat, indexNode.graphNodes[0].id)
		//}
	}

	var indexNodesCnt, GraphNodesCnt int
	var countIndexNodes func(*IndexNode)
	countIndexNodes = func(ro *IndexNode) {
		GraphNodesCnt++
		for _, n := range ro.next {
			countIndexNodes(n)
			indexNodesCnt += len(n.graphNodes)
		}
	}
	countIndexNodes(root)
	common.DebugLog(indexNodesCnt, GraphNodesCnt)

	var trimIndex func(*IndexNode)
	trimIndex = func(ro *IndexNode) {
		for k, n := range ro.next {
			trimIndex(n)
			if len(n.graphNodes) == 0 && len(n.next) == 0 {
				delete(ro.next, k)
			}
		}
	}
	trimIndex(root)

	indexNodesCnt = 0
	GraphNodesCnt = 0
	countIndexNodes(root)
	common.DebugLog(indexNodesCnt, GraphNodesCnt)

	return root
}

func SearchIndex(lat float64, lon float64, ro *IndexNode) *IndexNode {
	var searchIndex func(*common.GraphNode, *IndexNode, int64) *IndexNode
	searchIndex = func(gn *common.GraphNode, in *IndexNode, base int64) *IndexNode {
		nextKey := int64(gn.Lat*float64(base))*base*1000 + int64(gn.Lon*float64(base))
		//if gn.id == -1 {
		//	common.DebugLog("1111111111")
		//	common.DebugLog(gn.lat, gn.lon, base)
		//	common.DebugLog(nextKey)
		//}
		if in == nil || len(in.graphNodes) != 0 || len(in.next) == 0 {
			return in
		}
		//if gn.id == -1 {
		//	common.DebugLog(len(in.next[nextKey].next), len(in.next[nextKey].graphNodes))
		//}
		return searchIndex(gn, in.next[nextKey], base*10)
	}
	return searchIndex(&common.GraphNode{Lat: lat, Lon: lon}, ro, 1)
}

func PreComputing(graphNodes map[int64]*common.GraphNode) {
	var count, neighborCount int

	for _, startNode := range graphNodes {
		//common.InfoLog("precomputing ", count, len(graphNodes))
		count++
		neighborCount = 0

		pq := NewGraphHeap()
		pq.Push(NewHeapItem(startNode, 0, nil))

		startNode.ShortestPath[startNode] = &common.Path{
			StartPoint: startNode,
			EndPoint:   startNode,
			Points:     []*common.GraphNode{startNode},
			Distance:   0,
		}

		for pq.Len() > 0 && neighborCount < 50 {
			item := pq.Pop()
			currentNode := item.node
			currentDis := item.distance

			if _, exist := startNode.ShortestPath[currentNode]; currentNode != startNode && exist {
				continue
			}

			if currentNode != startNode {
				prevPath := startNode.ShortestPath[item.prevNode]

				newPath := &common.Path{
					StartPoint: startNode,
					EndPoint:   currentNode,
					Points:     make([]*common.GraphNode, 0, len(prevPath.Points)+1),
					Distance:   item.distance,
				}

				newPath.Points = append(newPath.Points, prevPath.Points...)
				newPath.Points = append(newPath.Points, currentNode)
				startNode.ShortestPath[currentNode] = newPath
				neighborCount++
			}

			for _, road := range currentNode.Next {
				nextNode := road.Node

				if _, exist := startNode.ShortestPath[nextNode]; exist {
					continue
				}

				newDis := currentDis + common.MAGIC_NUM*common.Distance(currentNode.Lat, currentNode.Lon, nextNode.Lat, nextNode.Lon)
				pq.Push(NewHeapItem(nextNode, newDis, currentNode))
			}
		}
	}
}

func PreProcess(points []*proto_struct.TrackPoint) map[string][]*proto_struct.TrackPoint {
	var curTime string
	groupedPoints := make(map[string][]*proto_struct.TrackPoint)
	for _, p := range points {
		k := fmt.Sprintf("%s-%d", p.Date, p.Hour)
		if k != curTime {
			curTime = k
			ps := make([]*proto_struct.TrackPoint, 0)
			ps = append(ps, p)
			groupedPoints[k] = ps
		} else {
			lastPoint := groupedPoints[k][len(groupedPoints[k])-1]
			if lastPoint.Latitude == p.Latitude && lastPoint.Longitude == p.Longitude {
				//common.DebugLog(lastPoint.Timestamp, p.Timestamp, "same")
				lastPoint.Speed = p.Speed
				lastPoint.CollectionTime = p.CollectionTime
				lastPoint.Timestamp = p.Timestamp
			} else {
				groupedPoints[k] = append(groupedPoints[k], p)
			}
		}
	}
	return groupedPoints
}

func StartReader(rch chan []*proto_struct.TrackPoint, wg *sync.WaitGroup) {
	defer wg.Done()
	basePath := common.REFINED_RAW_DATA_DIR_PATH

	var cnt1, cnt2 int
	endIndex := (common.BATCH + 1) * common.BATCH_SIZE
	for i := common.BATCH * common.BATCH_SIZE; i < endIndex; i++ {
		rows := pgpg.ReadTrackPointFromParquet(filepath.Join(basePath, fmt.Sprintf("%d.parquet", i)))
		if rows == nil || len(rows) == 0 {
			continue
		}
		cnt1 += len(rows)
		groupedRows := PreProcess(rows)
		for _, ps := range groupedRows {
			cnt2 += len(ps)
			rch <- ps
		}
		common.InfoLog("Finish reading ", i)
	}
	//for i, f := range fs {
	//	if filepath.Ext(f.Name()) == ".parquet" {
	//		rows := pgpg.ReadTrackPointFromParquet(filepath.Join(basePath, f.Name()))
	//		cnt1 += len(rows)
	//		groupedRows := PreProcess(rows)
	//		for _, ps := range groupedRows {
	//			cnt2 += len(ps)
	//			rch <- ps
	//		}
	//	}
	//if i == 100 {
	//	break
	//}
	//common.InfoLog("Finish reading ", i)
	//runtime.GC()
	//debug.FreeOSMemory()
	//}
	close(rch)
	common.InfoLog("read rows", cnt1)
	common.InfoLog("after preprocess", cnt2)
}

//func ReaderWorker(dirPath string, ch chan []*proto_struct.TrackPoint, wg *sync.WaitGroup) {
//	defer wg.Done()
//	dataFiles, err := os.ReadDir(dirPath)
//	if err != nil {
//		log.Fatal("Unable to read source path", err.Error(), dirPath)
//	}
//	//common.DebugLog("reader:", dirName1, dirName2, len(dataFiles))
//	for _, dataFile := range dataFiles {
//		//common.DebugLog("dataFile", dataFile.Name())
//		if filepath.Ext(dataFile.Name()) == ".csv" {
//			fName := filepath.Join(dirPath, dataFile.Name())
//			//common.DebugLog("start reader", fName)
//			points := CSV.ReadPointFromCSV(fName)
//			groupedPoints := PreProcess(points)
//			for _, ps := range groupedPoints {
//				ch <- ps
//			}
//		}
//	}
//	common.InfoLog("Finished reading ", dirPath)
//}

//func StartWriterManager(writerChan chan *proto_struct.Track, wg *sync.WaitGroup) {
//	defer wg.Done()
//
//	years := []int{2021, 2022}
//	month, day := 13, 32
//	register := make(map[int32]chan *proto_struct.Track)
//	//年
//	for y := 0; y < len(years); y++ {
//		//月
//		for m := 1; m < month; m++ {
//			//日
//			for d := 1; d < day; d++ {
//				t := years[y]*1e4 + m*1e2 + d
//				ch := make(chan *proto_struct.Track, 300)
//				register[int32(t)] = ch
//				//log.Println("Register ", t)
//				dirPath := filepath.Join(common.TRACK_DATA_DIR_PATH, fmt.Sprint(years[y]), fmt.Sprint(m), fmt.Sprint(d))
//				go WriterWorker(dirPath, ch, wg)
//			}
//		}
//	}
//
//	var cnt int
//	for {
//		t := <-writerChan
//		if t == nil {
//			for _, ch := range register {
//				close(ch)
//			}
//			common.InfoLog("Total track:", cnt)
//			return
//		}
//		var y, m, d, key int
//		if _, err := fmt.Sscanf(t.Date, "%d-%d-%d", &y, &m, &d); err != nil {
//			common.ErrorLog("error when reading date", err)
//		}
//		key = y*1e4 + m*1e2 + d
//		if _, ok := register[int32(key)]; !ok {
//			common.ErrorLog("error when write proto :invalid data", t.Date)
//		} else {
//			cnt++
//			if gcFrequency > 0 && cnt%gcFrequency == 0 {
//				runtime.GC()
//			}
//			register[int32(key)] <- t
//		}
//	}
//}
//
//func WriterWorker(dirPath string, wch chan *proto_struct.Track, wg *sync.WaitGroup) {
//	common.DebugLog("start writer worker for ", dirPath)
//	defer wg.Done()
//	for {
//		ms := <-wch
//		if ms == nil {
//			return
//		}
//		w := proto_tools.NewProtoBufWriter(filepath.Join(dirPath, fmt.Sprintf("%d_%d.protob", ms.Vin, ms.Tid)))
//		common.InfoLog("write file to ", ms.Vin, ms.Tid)
//		w.Write(ms)
//		w.Close()
//	}
//}

func StartWriterManager(writerChan chan *proto_struct.Track, wg *sync.WaitGroup) {
	defer wg.Done()

	register := make(map[int32]chan *proto_struct.Track)
	endIndex := (common.BATCH + 1) * common.BATCH_SIZE
	for i := common.BATCH * common.BATCH_SIZE; i < endIndex; i++ {
		ch := make(chan *proto_struct.Track, 100)
		register[int32(i)] = ch
		//log.Println("Register ", t)
		wg.Add(1)
		fPath := filepath.Join(common.TRACK_RAW_DATA_DIR_PATH, fmt.Sprint(i)+".prob")
		go WriterWorker(fPath, ch, wg)
	}

	var badTrackCnt1, badTrackCnt2, badPointCnt1, badPointCnt2 int
	file1, err := os.Create(filepath.Join(common.TRACK_RAW_DATA_DIR_PATH, "bad_track1.txt"))
	if err != nil {
		log.Println("error when create file", err)
	}
	defer file1.Close()
	bf1 := bufio.NewWriterSize(file1, 4*1024*1024)

	file2, err := os.Create(filepath.Join(common.TRACK_RAW_DATA_DIR_PATH, "bad_track2.txt"))
	if err != nil {
		log.Println("error when create file", err)
	}
	defer file2.Close()
	bf2 := bufio.NewWriterSize(file2, 4*1024*1024)

	var cnt int
	for {
		t := <-writerChan
		if t == nil {
			for _, ch := range register {
				close(ch)
			}
			_ = bf1.Flush()
			_ = bf2.Flush()
			common.InfoLog("Total track:", cnt)
			common.InfoLog("bad track1:", badTrackCnt1)
			common.InfoLog("bad track2:", badTrackCnt2)
			common.InfoLog("bad point cnt1:", badPointCnt1)
			common.InfoLog("bad point cnt2:", badPointCnt2)
			return
		}
		if t.IsBad == 1 {
			badTrackCnt1++
			badPointCnt1 += len(t.TrackSegs[0].TrackPoints)
			if _, err = bf1.WriteString(fmt.Sprintln(t.Date, t.Vin, t.Tid)); err != nil {
				common.ErrorLog("Error writing to bad track file", err)
			}
		} else if t.IsBad == 2 {
			badTrackCnt2++
			badPointCnt2 += len(t.TrackSegs[0].TrackPoints)
			if _, err = bf2.WriteString(fmt.Sprintln(t.Date, t.Vin, t.Tid)); err != nil {
				common.ErrorLog("Error writing to bad track file", err)
			}
		}
		if _, ok := register[t.Vin]; !ok {
			common.ErrorLog("error when write parquet :invalid vin", t.Vin)
		} else {
			cnt++
			if gcFrequency > 0 && cnt%gcFrequency == 0 {
				runtime.GC()
			}
			register[t.Vin] <- t
		}
	}

}

func WriterWorker(fPath string, wch chan *proto_struct.Track, wg *sync.WaitGroup) {
	w := proto_tools.NewProtoBufWriter(fPath)
	common.DebugLog("start writer worker for ", fPath)
	defer wg.Done()
	for {
		t := <-wch
		if t == nil {
			w.Close()
			return
		}
		w.Write(t)
	}
}

func ProcessData(readerChan chan []*proto_struct.TrackPoint, writerChan chan *proto_struct.Track, indexRoot *IndexNode) {
	wg := &sync.WaitGroup{}

	for i := 1; i < workerCount; i++ {
		wg.Add(1)
		go Worker(readerChan, writerChan, indexRoot, wg)
	}
	wg.Add(1)
	Worker(readerChan, writerChan, indexRoot, wg)

	//runtime.GC()
	////禁用自动GC
	//debug.SetGCPercent(-1)

	wg.Wait()
	close(writerChan)
}

func Worker(readerChan chan []*proto_struct.TrackPoint, writerChan chan *proto_struct.Track, indexRoot *IndexNode, wg *sync.WaitGroup) {
	defer wg.Done()
	var cnt int
	var candiTime, pathTime float64
	for {
		ps := <-readerChan
		cnt++

		st := time.Now()
		if ps == nil {
			//common.InfoLog("time for searching candidate", candiTime)
			//common.InfoLog("time for searching path", pathTime)
			//common.InfoLog("processed track", cnt)
			return
		}

		var firstCandi, curCandi *common.CandidateSet
		mmap := common.NewMemoryMap()
		for _, p := range ps {
			mmap.RecordTrackPoint(p)
			//121.664006 31.160551
			//common.DebugLog("point:", p.Longitude, p.Latitude)
			//以目标点为中心的九宫格
			directions := []float64{-0.01, 0, 0.01}
			ins := make([]*IndexNode, 0, 9)
			for _, d1 := range directions {
				for _, d2 := range directions {
					in := SearchIndex(p.Latitude+d1, p.Longitude+d2, indexRoot)
					if in != nil {
						ins = append(ins, in)
					}
				}
			}

			//if len(ins) == 0 {
			//	common.DebugLog("==========1===========")
			//	common.DebugLog(p.Latitude, p.Longitude)
			//}

			cs := CandidateSearch(p, ins, mmap)

			candiTime += time.Since(st).Seconds()
			//common.InfoLog("time for candidate search: ", time.Since(startTime))
			st = time.Now()

			//if cs == nil {
			//	common.DebugLog("=======2==============")
			//	common.DebugLog(p.Latitude, p.Longitude)
			//}

			if curCandi == nil {
				curCandi = &common.CandidateSet{Cp: cs}
				firstCandi = curCandi
			} else {
				curCandi.Next = &common.CandidateSet{Cp: cs}
				curCandi = curCandi.Next
			}
			//printCandidate(curCandi.Cp)
		}

		if firstCandi == nil {
			common.DebugLog("candidate set is empty")
			common.DebugLog("=================")
			for _, p := range ps {
				common.DebugLog(p.Vin, p.Hour, p.Date, p.Timestamp, p.CollectionTime, p.Speed, p.Latitude, p.Longitude)
			}
			common.DebugLog("=================")
			return
		}

		var discreteCount int
		var res *proto_struct.Track
		for c := firstCandi; c != nil; c = c.Next {
			if len(c.Cp) == 1 && c.Cp[0].Ttype == common.DISCRETE {
				discreteCount++
			}
		}
		//common.DebugLog("discrete count", discreteCount, "total count", len(ps))
		if float64(discreteCount)/float64(len(ps)) > 0.5 {
			//common.InfoLog("bad track", firstCandi.Cp[0].originalPoint.Date, firstCandi.Cp[0].originalPoint.Vin, firstCandi.Cp[0].originalPoint.Hour, firstCandi.Cp[0].originalPoint.Hour, firstCandi.Cp[0].originalPoint.Timestamp)
			res = common.GetTrack(mmap)
			res.Vin = ps[0].Vin
			res.Tid = ps[0].Hour
			res.StartTime = ps[0].Timestamp
			res.EndTime = ps[len(ps)-1].Timestamp
			res.Date = ps[0].Date
			res.TrackSegs = []*proto_struct.TrackSegment{common.GetTrackSegment(mmap)}
			res.TrackSegs[0].TrackPoints = ps
			//res.OriginalPoints = ps
			res.IsBad = 1
		} else {
			candidateTracks := make(map[*common.CandidatePoint]*proto_struct.Track)
			for _, p := range firstCandi.Cp {
				t := BuildTrack(p, mmap)
				candidateTracks[p] = t
			}
			for c := firstCandi.Next; c != nil; c = c.Next {
				//if len(candidateTracks) > 0 {
				//	for p := range candidateTracks {
				//		common.DebugLog("point:", p.lat, p.lon, p.originalPoint.Latitude, p.originalPoint.Longitude)
				//	}
				//}
				//common.DebugLog("candidateTracks:", len(c.Cp), len(candidateTracks))
				candidateTracks = FindOptimalPath(c, candidateTracks, mmap)
				if len(candidateTracks) == 0 {
					//common.InfoLog("bad track", c.Cp[0].originalPoint.Date, c.Cp[0].originalPoint.Vin, c.Cp[0].originalPoint.Hour, c.Cp[0].originalPoint.Timestamp)
					res = common.GetTrack(mmap)
					res.Vin = ps[0].Vin
					res.Tid = ps[0].Hour
					res.StartTime = ps[0].Timestamp
					res.EndTime = ps[len(ps)-1].Timestamp
					res.Date = ps[0].Date
					res.TrackSegs = []*proto_struct.TrackSegment{common.GetTrackSegment(mmap)}
					res.TrackSegs[0].TrackPoints = ps
					res.IsBad = 2
					//OriginalPoints: ps,
					break
				}
			}

			if len(candidateTracks) != 0 {
				var maxProb float64 = -1
				for _, t := range candidateTracks {
					if maxProb < t.Probability {
						maxProb = t.Probability
						res = t
						//common.DebugLog("result is ", res.Probability)
					}
				}
				//common.InfoLog("send to writer ", res.Vin, res.Tid, res.Date, res.StartTime, res.EndTime, res.Probability)
			}
		}
		if res == nil {
			common.InfoLog("no track found", len(ps), ps[0].Vin, ps[0].Hour, ps[0].Date, ps[0].Timestamp)
		}
		res = DeepCopyTrack(res)
		mmap.Clear()
		pathTime += time.Since(st).Seconds()
		writerChan <- res
	}
}

func CandidateSearch(p *proto_struct.TrackPoint, ins []*IndexNode, mmp *common.MemoryMap) []*common.CandidatePoint {
	var minNode, minNode2 *common.GraphNode
	minDis := math.MaxFloat64
	mint := math.MaxFloat64
	var minID int64
	var minLat, minLon float64
	targetMap := make(map[int64]int64)
	candidates := make([]*common.CandidatePoint, 0)
	for _, in := range ins {
		for _, gn := range in.graphNodes {
			for _, way := range gn.Next {
				//双向路会产生两个candidate，去重
				if k, exist := targetMap[gn.Id]; exist && k == way.Node.Id {
					continue
				}
				x1, y1, x2, y2, x3, y3 := gn.Lat, gn.Lon, way.Node.Lat, way.Node.Lon, p.Latitude, p.Longitude
				t := common.CalT(x1, y1, x2, y2, x3, y3)
				//if gn.id == 475432841 || way.Node.id == 475432841 {
				//	common.DebugLog("t:", t)
				//}
				var dis float64
				if t < 1 && t > 0 {
					dis = common.P2lDistance(x1, y1, x2, y2, x3, y3) * common.MAGIC_NUM
					//if t < 0 {
					//	t = 0
					//}
					//所有允许范围内的值
					if dis < DistanceOffset {
						cand := common.GetCandidatePoint(mmp)
						cand.Vertex = []*common.GraphNode{gn, way.Node}
						cand.Ttype = common.NORMAL
						cand.TT = t
						cand.Distance = dis
						cand.OriginalPoint = p
						targetMap[way.Node.Id] = gn.Id
						//common.DebugLog("dis:", dis)
						cand.Lat, cand.Lon = common.CalP(x1, x2, y1, y2, t)
						cand.Ep = common.CalEP(dis)
						cand.RoadID = way.ID
						candidates = append(candidates, cand)
						//common.DebugLog(cand.Lat, cand.Lon)
					}
					//记录最小值
					if dis < minDis {
						mint = t
						minNode = gn
						minNode2 = way.Node
						minDis = dis
						minID = way.ID
						minLat, minLon = common.CalP(x1, x2, y1, y2, t)
					}
				} else {
					//如果不垂直检查端点距离
					dis = common.Distance(gn.Lat, gn.Lon, p.Latitude, p.Longitude) * common.MAGIC_NUM
					if dis < DistanceOffset {
						cand := common.GetCandidatePoint(mmp)
						cand.Vertex = []*common.GraphNode{gn, way.Node}
						cand.Ttype = common.NORMAL
						cand.TT = t
						cand.Distance = dis
						cand.OriginalPoint = p
						cand.Lat = gn.Lat
						cand.Lon = gn.Lon
						targetMap[way.Node.Id] = gn.Id
						//common.DebugLog("dis:", dis)
						cand.Ep = common.CalEP(dis)
						cand.RoadID = way.ID
						candidates = append(candidates, cand)
						//common.DebugLog(cand.lat, cand.lon)
					}
					//记录最小值
					if dis < minDis {
						mint = t
						minNode = gn
						minNode2 = way.Node
						minDis = dis
						minID = way.ID
						minLat = gn.Lat
						minLon = gn.Lon
					}
				}
			}
		}
	}
	if len(candidates) == 0 {
		cp := common.GetCandidatePoint(mmp)
		cp.Vertex = []*common.GraphNode{minNode, minNode2}
		cp.Ttype = common.DISCRETE
		cp.TT = mint
		cp.Distance = minDis
		cp.RoadID = minID
		cp.Ep = 1
		cp.OriginalPoint = p
		cp.Lat = minLat
		cp.Lon = minLon
		return []*common.CandidatePoint{cp}
	}
	return candidates
}

func printCandidate(cp []*common.CandidatePoint) {
	for _, x := range cp {
		common.DebugLog("candidates:")
		common.DebugLog(x.Vertex[0].Lat, x.Vertex[0].Lon, x.Vertex[1].Lat, x.Vertex[1].Lon)
		common.DebugLog(x.OriginalPoint)
		common.DebugLog(x.Distance)
		common.DebugLog(x.TT)
		common.DebugLog(x.Lat, x.Lon)
	}
}

//func BuildTrack(p *CandidatePoint) *proto_struct.Track {
//	tp := &proto_struct.TrackPoint{
//		Vin:            p.originalPoint.Vin,
//		CollectionTime: p.originalPoint.CollectionTime,
//		Date:           p.originalPoint.Date,
//		Timestamp:      p.originalPoint.Timestamp,
//		Hour:           p.originalPoint.Hour,
//		Speed:          p.originalPoint.Speed,
//		Longitude:      p.lon,
//		Latitude:       p.Lat,
//	}
//	if p.Ttype == DISCRETE {
//		tp.Longitude = p.originalPoint.Longitude
//		tp.Latitude = p.originalPoint.Latitude
//	}
//	return &proto_struct.Track{
//		Vin:            p.originalPoint.Vin,
//		Tid:            p.originalPoint.Hour,
//		Date:           p.originalPoint.Date,
//		StartTime:      p.originalPoint.Timestamp,
//		EndTime:        p.originalPoint.Timestamp,
//		OriginalPoints: []*proto_struct.TrackPoint{p.originalPoint},
//		TrackPoints:    []*proto_struct.TrackPoint{tp},
//		TrackSegs: []*proto_struct.TrackSegment{
//			{
//				StartTime: p.originalPoint.Timestamp,
//				EndTime:   p.originalPoint.Timestamp,
//				RoadID:    p.roadID,
//			},
//		},
//		Probability: p.ep,
//	}
//}

func BuildTrack(p *common.CandidatePoint, mmp *common.MemoryMap) *proto_struct.Track {
	tp := common.GetTrackPoint(mmp)
	tp.Vin = p.OriginalPoint.Vin
	tp.CollectionTime = p.OriginalPoint.CollectionTime
	tp.Date = p.OriginalPoint.Date
	tp.Timestamp = p.OriginalPoint.Timestamp
	tp.Hour = p.OriginalPoint.Hour
	tp.Speed = p.OriginalPoint.Speed
	tp.Longitude = p.Lon
	tp.Latitude = p.Lat
	if p.Ttype == common.DISCRETE {
		tp.Longitude = p.OriginalPoint.Longitude
		tp.Latitude = p.OriginalPoint.Latitude
	}

	t := common.GetTrack(mmp)

	ts := common.GetTrackSegment(mmp)
	ts.StartTime = p.OriginalPoint.Timestamp
	ts.EndTime = p.OriginalPoint.Timestamp
	ts.RoadId = p.RoadID
	ts.OriginalPoints = []*proto_struct.TrackPoint{p.OriginalPoint}
	ts.TrackPoints = []*proto_struct.TrackPoint{tp}
	t.TrackSegs = []*proto_struct.TrackSegment{ts}

	t.Vin = p.OriginalPoint.Vin
	t.Tid = p.OriginalPoint.Hour
	t.Date = p.OriginalPoint.Date
	t.StartTime = p.OriginalPoint.Timestamp
	t.EndTime = p.OriginalPoint.Timestamp
	t.Probability = p.Ep
	return t
}

//	func AppendTrack(t *proto_struct.Track, p *common.CandidatePoint, pa *common.Path) {
//		t.EndTime = p.originalPoint.Timestamp
//		t.OriginalPoints = append(t.OriginalPoints, p.originalPoint)
//		//t.Probability = p.probability
//
//		if pa != nil {
//			for _, gp := range pa.points {
//				t.TrackPoints = append(t.TrackPoints, &proto_struct.TrackPoint{
//					Vin:       p.originalPoint.Vin,
//					Speed:     p.originalPoint.Speed,
//					Timestamp: p.originalPoint.Timestamp,
//					Date:      p.originalPoint.Date,
//					Hour:      p.originalPoint.Hour,
//					Longitude: gp.lon,
//					Latitude:  gp.Lat,
//				})
//				if len(t.TrackSegs) == 0 || t.TrackSegs[len(t.TrackSegs)-1].RoadID != p.roadID {
//					if len(t.TrackSegs) != 0 {
//						t.TrackSegs[len(t.TrackSegs)-1].EndTime = p.originalPoint.Timestamp
//					}
//					t.TrackSegs = append(t.TrackSegs, &proto_struct.TrackSegment{
//						StartTime: p.originalPoint.Timestamp,
//						EndTime:   p.originalPoint.Timestamp,
//						RoadID:    p.roadID,
//					})
//				} else {
//					t.TrackSegs[len(t.TrackSegs)-1].EndTime = p.originalPoint.Timestamp
//				}
//			}
//		}
//
//		tp := &proto_struct.TrackPoint{
//			Vin:       p.originalPoint.Vin,
//			Hour:      p.originalPoint.Hour,
//			Timestamp: p.originalPoint.Timestamp,
//			Latitude:  p.Lat,
//			Longitude: p.lon,
//		}
//		t.TrackPoints = append(t.TrackPoints, tp)
//		//同一条路的后继离散点直接记录原始点
//		if p.Ttype == DISCRETE {
//			tp.Longitude = p.originalPoint.Longitude
//			tp.Latitude = p.originalPoint.Latitude
//		}
//
// }
func AppendTrack(t *proto_struct.Track, p *common.CandidatePoint, pa *common.Path, mmp *common.MemoryMap) {
	t.EndTime = p.OriginalPoint.Timestamp
	currentSeg := t.TrackSegs[len(t.TrackSegs)-1]
	if currentSeg.RoadId != p.RoadID {
		currentSeg = common.GetTrackSegment(mmp)
		currentSeg.StartTime = p.OriginalPoint.Timestamp
		currentSeg.EndTime = p.OriginalPoint.Timestamp
		currentSeg.RoadId = p.RoadID
		t.TrackSegs = append(t.TrackSegs, currentSeg)
	}
	currentSeg.OriginalPoints = append(currentSeg.OriginalPoints, p.OriginalPoint)
	currentSeg.EndTime = p.OriginalPoint.Timestamp
	//t.Probability = p.probability

	if pa != nil {
		for _, gp := range pa.Points {

			ntp := common.GetTrackPoint(mmp)
			ntp.Vin = p.OriginalPoint.Vin
			ntp.Speed = p.OriginalPoint.Speed
			ntp.Timestamp = p.OriginalPoint.Timestamp
			ntp.Date = p.OriginalPoint.Date
			ntp.Hour = p.OriginalPoint.Hour
			ntp.Longitude = gp.Lon
			ntp.Latitude = gp.Lat
			ntp.CollectionTime = p.OriginalPoint.CollectionTime

			currentSeg.TrackPoints = append(currentSeg.TrackPoints, ntp)
		}
	}

	tp := common.GetTrackPoint(mmp)
	tp.Vin = p.OriginalPoint.Vin
	tp.Hour = p.OriginalPoint.Hour
	tp.Timestamp = p.OriginalPoint.Timestamp
	tp.CollectionTime = p.OriginalPoint.CollectionTime
	tp.Latitude = p.Lat
	tp.Longitude = p.Lon

	currentSeg.TrackPoints = append(currentSeg.TrackPoints, tp)
	//同一条路的后继离散点直接记录原始点
	//if p.Ttype == common.DISCRETE {
	//	tp.Longitude = p.OriginalPoint.Longitude
	//	tp.Latitude = p.OriginalPoint.Latitude
	//}

}

func CopyTrack(t *proto_struct.Track, mmp *common.MemoryMap) *proto_struct.Track {
	newTrack := common.GetTrack(mmp)
	newTrack.Vin = t.Vin
	newTrack.Tid = t.Tid
	newTrack.StartTime = t.StartTime
	newTrack.EndTime = t.EndTime
	newTrack.Date = t.Date
	newTrack.Probability = t.Probability

	//newTrack.TrackSegs = append(newTrack.TrackSegs, t.TrackSegs...)

	lastIndex := len(t.TrackSegs) - 1
	if len(t.TrackSegs) > 1 {
		newTrack.TrackSegs = append(newTrack.TrackSegs, t.TrackSegs[:lastIndex]...)
	}

	lastSeg := t.TrackSegs[lastIndex]
	newSeg := common.GetTrackSegment(mmp)
	newSeg.StartTime = lastSeg.StartTime
	newSeg.EndTime = lastSeg.EndTime
	//newSeg.OriginalPoints = make([]*proto_struct.TrackPoint, 0, len(seg.OriginalPoints))
	newSeg.OriginalPoints = lastSeg.OriginalPoints
	//newSeg.TrackPoints = make([]*proto_struct.TrackPoint, 0, len(seg.TrackPoints))
	newSeg.TrackPoints = lastSeg.TrackPoints

	newTrack.TrackSegs = append(newTrack.TrackSegs, newSeg)
	return newTrack
}

func DeepCopyTrack(t *proto_struct.Track) *proto_struct.Track {
	newTrack := common.GetTrack(nil)
	newTrack.Vin = t.Vin
	newTrack.Tid = t.Tid
	newTrack.StartTime = t.StartTime
	newTrack.EndTime = t.EndTime
	newTrack.Date = t.Date
	newTrack.Probability = t.Probability

	for _, ts := range t.TrackSegs {
		newTs := common.GetTrackSegment(nil)
		newTs.StartTime = ts.StartTime
		newTs.EndTime = ts.EndTime
		newTs.RoadId = ts.RoadId
		newTrack.TrackSegs = append(newTrack.TrackSegs, newTs)

		for _, op := range ts.OriginalPoints {
			tp := common.GetTrackPoint(nil)
			tp.Vin = op.Vin
			tp.Hour = op.Hour
			tp.Longitude = op.Longitude
			tp.Latitude = op.Latitude
			tp.Date = op.Date
			tp.Speed = op.Speed
			tp.Timestamp = op.Timestamp
			tp.CollectionTime = op.CollectionTime
			newTs.OriginalPoints = append(newTs.OriginalPoints, tp)
		}
		for _, op := range ts.TrackPoints {
			tp := common.GetTrackPoint(nil)
			tp.Vin = op.Vin
			tp.Hour = op.Hour
			tp.Longitude = op.Longitude
			tp.Latitude = op.Latitude
			tp.Date = op.Date
			tp.Speed = op.Speed
			tp.Timestamp = op.Timestamp
			tp.CollectionTime = op.CollectionTime
			newTs.TrackPoints = append(newTs.TrackPoints, tp)
		}
	}

	return newTrack
}

func OnSameRoad(p1, p2 *common.CandidatePoint) bool {
	if p1 == nil || p2 == nil {
		return false
	}
	if p1.Vertex == nil || p2.Vertex == nil || len(p1.Vertex) != len(p2.Vertex) || len(p1.Vertex) != 2 {
		return false
	}
	return (p1.Vertex[0] == p2.Vertex[0] && p1.Vertex[1] == p2.Vertex[1]) || (p1.Vertex[0] == p2.Vertex[1] && p1.Vertex[1] == p2.Vertex[0])
}

func FindOptimalPath(nextCandis *common.CandidateSet, tracks map[*common.CandidatePoint]*proto_struct.Track, mmap *common.MemoryMap) map[*common.CandidatePoint]*proto_struct.Track {
	res := make(map[*common.CandidatePoint]*proto_struct.Track)
	for _, n := range nextCandis.Cp {
		//common.DebugLog("search for", n.originalPoint.Latitude, n.originalPoint.Longitude)
		//common.DebugLog("candis", n.Lat, n.lon, n.Distance, n.roadID)
		var maxScore float64 = -1
		var maxTrack *proto_struct.Track
		var maxPath *common.Path
		for p, t := range tracks {
			//common.DebugLog("p", p.originalPoint.Latitude, p.originalPoint.Longitude)
			//common.DebugLog("pp", p.Lat, p.lon, p.roadID)
			//common.DebugLog(p.originalPoint.Vin, p.originalPoint.Date, p.originalPoint.Timestamp, p.originalPoint.Hour)
			ddistance := common.Distance(p.OriginalPoint.Latitude, p.OriginalPoint.Longitude, n.OriginalPoint.Latitude, n.OriginalPoint.Longitude) * common.MAGIC_NUM
			ldistance := math.MaxFloat64
			var sp *common.Path

			//如果是在同一段路上
			//if ((p.Ttype == DISCRETE || n.Ttype == DISCRETE) && ddistance < 200) || p.Ttype == DISCRETE && n.Ttype == DISCRETE {
			if p.Ttype == common.DISCRETE || n.Ttype == common.DISCRETE {
				ldistance = ddistance
			} else if OnSameRoad(p, n) == true {
				//common.DebugLog("same road")
				ldistance = common.Distance(p.Lat, p.Lon, n.Lat, n.Lon) * common.MAGIC_NUM
			} else {
				var startNode, endNode *common.GraphNode
				//common.DebugLog("search road1:", p.Vertex[0].Lat, p.Vertex[0].lon)
				//common.DebugLog("search road2:", p.Vertex[1].Lat, p.Vertex[1].lon)
				//common.DebugLog("search road3:", n.Vertex[0].Lat, n.Vertex[0].lon)
				//common.DebugLog("search road4:", n.Vertex[1].Lat, n.Vertex[1].lon)

				for i := 0; i < len(p.Vertex); i++ {
					for j := 0; j < len(n.Vertex); j++ {
						ld := common.Distance(p.Lat, p.Lon, p.Vertex[i].Lat, p.Vertex[i].Lon)*common.MAGIC_NUM + common.Distance(n.Lat, n.Lon, n.Vertex[j].Lat, n.Vertex[j].Lon)*common.MAGIC_NUM
						//common.DebugLog(p.Lat, p.lon, p.Vertex[i].Lat, p.Vertex[i].lon)
						//common.DebugLog(n.Lat, n.lon, n.Vertex[j].Lat, n.Vertex[j].lon)
						//common.DebugLog(p.Vertex[i].id, n.Vertex[j].id)
						//common.DebugLog("ld:", ld)
						if pa, exist := p.Vertex[i].ShortestPath[n.Vertex[j]]; exist {
							ld += pa.Distance
							if ldistance > ld {
								ldistance = ld
								startNode = p.Vertex[i]
								endNode = n.Vertex[j]
								sp = pa
							}
						} else {
							//不存在说明距离过远
							//ddistance = math.MaxFloat64
							//common.DebugLog("not exist")
						}
					}
				}
				if startNode == nil {
					//无法连通说明此候选点应当排除
					//common.DebugLog("not exist2")
					continue
				}
				common.DebugLog(startNode.Lat, startNode.Lon, endNode.Lat, endNode.Lon)
			}

			score := t.Probability + (min(ddistance, ldistance)/max(ddistance, ldistance))*n.Ep
			//common.DebugLog("ddistance", ddistance)
			//common.DebugLog("ldistance", ldistance)
			//common.DebugLog("probability", newT.Probability)
			if score > maxScore {
				maxScore = score
				maxTrack = t
				maxPath = sp
			}
		}
		if maxTrack != nil {
			newT := CopyTrack(maxTrack, mmap)
			AppendTrack(newT, n, maxPath, mmap)
			res[n] = newT
		}
	}
	return res
}
