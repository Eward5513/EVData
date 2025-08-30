package main

import (
	"EVdata/CSV"
	"EVdata/common"
	"EVdata/pgpg"
	"EVdata/proto_struct"
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
	workerCount    int
	DistanceOffset float64
)

func main() {

	begin := time.Now()

	var profMode bool
	flag.BoolVar(&profMode, "p", false, "pprof mode")
	flag.Float64Var(&DistanceOffset, "d", 40, "Distance offset")
	flag.IntVar(&workerCount, "wc", 1000, "worker count")
	flag.Parse()

	common.CreatLogFile("mapmatching.log")
	defer common.CloseLogFile()

	if profMode {
		//common.SetLogLevel(common.DEBUG)
		go func() {
			log.Println("start pprof tool")
			//http://localhost:6060/debug/pprof/
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	common.SetLogLevel(common.INFO)

	t := time.Now()
	graph := BuildGraph("shanghai_new.json")
	common.InfoLog("time for building graph: ", time.Since(t))
	t = time.Now()

	indexRoot := BuildIndex(graph)
	common.InfoLog("time for building index: ", time.Since(t))
	t = time.Now()

	PreComputing(graph)
	common.InfoLog("time for precomputing: ", time.Since(t))

	ProcessDataLoop(indexRoot)

	log.Println("Total Execution time: ", time.Now().Sub(begin))
}

func ProcessDataLoop(indexRoot *IndexNode) {
	readerChan := make(chan []*proto_struct.RawPoint, 1000)
	writerChan := make(chan *proto_struct.Track, 1000)

	go StartReader(readerChan)

	go ProcessData(readerChan, writerChan, indexRoot)

	StartWriterWorker(writerChan)
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
	graphNodes := make(map[int64]*common.GraphNode)

	for _, element := range roadNetwork.Elements {
		if element.Type == "way" {
			roads[element.ID] = &element
		} else if element.Type == "node" {

			common.MIN_LONGITUDE = min(common.MIN_LONGITUDE, element.Lon)
			common.MIN_LATITUDE = min(common.MIN_LATITUDE, element.Lat)
			common.MAX_LATITUDE = max(common.MAX_LATITUDE, element.Lat)
			common.MAX_LONGITUDE = max(common.MAX_LONGITUDE, element.Lon)

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
		indexNode := SearchIndex(n.Lat, n.Lon, root)
		indexNode.graphNodes = append(indexNode.graphNodes, n)
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
		if in == nil || len(in.graphNodes) != 0 || len(in.next) == 0 {
			return in
		}
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

func StartReader(rch chan []*proto_struct.RawPoint) {
	parquetPath := common.RAW_POINT_PARQUET_PATH

	vps := pgpg.ReadPointFromParquet(parquetPath)

	common.InfoLog("Finish reading parquet file")

	var cnt int
	//for i := 1; i < len(vps); i++ {
	//	cnt += len(vps[i])
	//	if i%1000 == 0 {
	//		common.InfoLog("Finish reading ", i, " / ", len(vps))
	//	}
	//	rch <- vps[i]
	//}
	rch <- vps[1]
	close(rch)
	common.InfoLog("Finish reading ", cnt)
}

func StartWriterWorker(writerChan chan *proto_struct.Track) {
	wg := &sync.WaitGroup{}
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			for t := range writerChan {
				if len(t.Mps) == 0 {
					common.InfoLog("data is empty")
					continue
				}
				bw := CSV.NewMatchingPointWriter(filepath.Join(common.MATCHED_POINT_CSV_DIR, fmt.Sprintf("%d_%d.csv", t.Vin, t.Tid)))
				bw.Write(t.Mps)
				//common.InfoLog("Finish writing ", t.Vin, t.Tid)
				common.PutTrack(t)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func ProcessData(readerChan chan []*proto_struct.RawPoint, writerChan chan *proto_struct.Track, indexRoot *IndexNode) {

	wg := &sync.WaitGroup{}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go Worker(readerChan, writerChan, indexRoot, wg)
	}

	wg.Wait()
	close(writerChan)
}

func Worker(readerChan chan []*proto_struct.RawPoint, writerChan chan *proto_struct.Track, indexRoot *IndexNode, wg *sync.WaitGroup) {
	defer wg.Done()

	mmap := common.NewMemoryMap()

	for msg := range readerChan {
		//common.InfoLog("worker read", len(msg))

		mmap.RecordRawPoints(msg)

		tps := PreProcess(msg)

		for tid, ps := range tps {
			common.InfoLog(tid, len(ps), ps[0].Time)

			var firstCandi, curCandi *common.CandidateSet
			for _, p := range ps {
				cs := CandidateSearch(p, indexRoot, mmap, 0)

				if curCandi == nil {
					curCandi = &common.CandidateSet{Cp: cs}
					firstCandi = curCandi
				} else {
					curCandi.Next = &common.CandidateSet{Cp: cs}
					curCandi = curCandi.Next
				}
			}

			if firstCandi == nil {
				common.InfoLog("candidate set is empty")
				continue
			}

			var res *proto_struct.Track

			candidateTracks := make(map[*common.CandidatePoint]*proto_struct.Track)
			for _, p := range firstCandi.Cp {
				t := BuildTrack(tid, p, mmap)
				candidateTracks[p] = t
			}
			var retry int
			for c := firstCandi.Next; c != nil; c = c.Next {
				bak := candidateTracks
				candidateTracks = FindOptimalPath(c, candidateTracks, mmap)
				if len(candidateTracks) == 0 {
					//第一次尝试扩大搜索范围
					if retry == 0 {
						retry++
						cps := CandidateSearch(c.Cp[0].OriginalPoint, indexRoot, mmap, float64(retry))
						c.Next = &common.CandidateSet{Cp: cps, Next: c.Next}
						candidateTracks = bak
					} else if retry == 1 {
						//第二次将该点标记为离散点
						retry++
						cand := mmap.GetCandidatePoint()
						cand.OriginalPoint = c.Cp[0].OriginalPoint
						cand.Ttype = common.DISCRETE
						c.Next = &common.CandidateSet{Cp: []*common.CandidatePoint{cand}, Next: c.Next}
						candidateTracks = bak
					}
				} else {
					retry = 0
				}

			}

			if len(candidateTracks) != 0 {
				var maxProb float64 = -1
				for _, t := range candidateTracks {
					if maxProb < t.Probability {
						maxProb = t.Probability
						res = t
						//common.InfoLog("result is ", res.Probability)
					}
				}
				//common.InfoLog("send to writer ", res.Vin, res.Tid, res.Date, res.StartTime, res.EndTime, res.Probability)
			}
			if res == nil {
				common.InfoLog("no track found")
				continue
			}

			ans := DeepCopyTrack(res)
			//common.InfoLog("send to writer", len(ans.Mps), ans.Vin, ans.Tid)
			mmap.Clear()
			writerChan <- ans
		}
	}
}

func PreProcess(msg []*proto_struct.RawPoint) [][]*proto_struct.RawPoint {
	var curTime int64 = -1
	var res [][]*proto_struct.RawPoint
	for _, p := range msg {
		//common.InfoLog(curTime, p.TimeInt)
		if curTime == -1 || p.TimeInt-curTime > 60 {
			tp := []*proto_struct.RawPoint{p}
			res = append(res, tp)
		} else {
			lastPs := res[len(res)-1]
			lastP := lastPs[len(lastPs)-1]
			if lastP.Longitude != p.Longitude && lastP.Latitude != p.Latitude {
				lastPs = append(lastPs, p)
				res[len(res)-1] = lastPs
			}
		}
		curTime = p.TimeInt
	}
	return res
}

func CandidateSearch(p *proto_struct.RawPoint, indexRoot *IndexNode, mmp *common.MemoryMap, tt float64) []*common.CandidatePoint {
	var minNode, minNode2 *common.GraphNode
	minDis := math.MaxFloat64
	mint := math.MaxFloat64
	var minID int64
	var minLat, minLon float64
	targetMap := make(map[int64]int64)
	candidates := make([]*common.CandidatePoint, 0)
	do := DistanceOffset + tt*DistanceOffset

	//以目标点为中心的9/16宫格
	var directions []float64
	if tt == 0 {
		directions = []float64{-0.01, 0, 0.01}
	} else {
		directions = []float64{-0.02, -0.01, 0, 0.01, 0.02}
	}
	ins := make([]*IndexNode, 0)
	for _, d1 := range directions {
		for _, d2 := range directions {
			in := SearchIndex(p.Latitude+d1, p.Longitude+d2, indexRoot)
			if in != nil {
				ins = append(ins, in)
			}
		}
	}
	visited := make(map[float64]float64)

	for _, in := range ins {
		for _, gn := range in.graphNodes {
			for _, way := range gn.Next {
				//双向路会产生两个candidate，去重
				if k, exist := targetMap[gn.Id]; exist && k == way.Node.Id {
					continue
				}
				x1, y1, x2, y2, x3, y3 := gn.Lat, gn.Lon, way.Node.Lat, way.Node.Lon, p.Latitude, p.Longitude
				t := common.CalT(x1, y1, x2, y2, x3, y3)
				var dis float64
				if t < 1 && t > 0 {
					dis = common.P2lDistance(x1, y1, x2, y2, x3, y3) * common.MAGIC_NUM
					//if t < 0 {
					//	t = 0
					//}
					//所有允许范围内的值
					if dis < do {
						lat, lon := common.CalP(x1, x2, y1, y2, t)
						if v, e := visited[lat]; e == false || v != lon {
							cand := mmp.GetCandidatePoint()
							cand.Vertex = []*common.GraphNode{gn, way.Node}
							cand.Ttype = common.NORMAL
							cand.TT = t
							cand.Distance = dis
							cand.OriginalPoint = p
							targetMap[way.Node.Id] = gn.Id
							//common.DebugLog("dis:", dis)
							cand.Lat, cand.Lon = lat, lon
							cand.Ep = common.CalEP(dis)
							cand.RoadID = way.ID
							candidates = append(candidates, cand)
							//if gn.Id == 6697251679 && p.Date == "2022-03-19" && p.Timestamp == "03:49:15" {
							//	common.InfoLog(gn.Id, way.Node.Id, cand.Lat, cand.Lon)
							//}
							visited[lat] = lon
						}
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
					if dis < do {
						if v, e := visited[gn.Lat]; e == false || v != gn.Lon {
							cand := mmp.GetCandidatePoint()
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
							visited[gn.Lat] = gn.Lon
						}
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
		cp := mmp.GetCandidatePoint()
		cp.Vertex = []*common.GraphNode{minNode, minNode2}
		cp.Ttype = common.DISCRETE
		cp.TT = mint
		cp.Distance = minDis
		cp.RoadID = minID
		cp.Ep = 1
		cp.OriginalPoint = p
		cp.Lat = minLat // 范围内搜不到的话会是零值
		cp.Lon = minLon
		return []*common.CandidatePoint{cp}
	}
	return candidates
}

func BuildTrack(tid int, p *common.CandidatePoint, mmap *common.MemoryMap) *proto_struct.Track {

	t := mmap.GetTrack()
	t.Vin = p.OriginalPoint.Vin
	t.Probability = 0
	t.Tid = int32(tid)

	mp := mmap.GetMatchingPoint()
	mp.RoadId = p.RoadID
	mp.OriginalLon = p.OriginalPoint.Longitude
	mp.OriginalLat = p.OriginalPoint.Latitude
	mp.MatchedLon = p.Lon
	mp.MatchedLat = p.Lat
	if p.Ttype == common.DISCRETE {
		mp.IsBad = 1
	}
	t.Mps = append(t.Mps, mp)

	return t
}

func AppendTrack(t *proto_struct.Track, p *common.CandidatePoint, pa *common.Path, mmap *common.MemoryMap) {
	if pa != nil {
		for _, gp := range pa.Points {
			mp := mmap.GetMatchingPoint()
			mp.OriginalLon = p.OriginalPoint.Longitude
			mp.OriginalLat = p.OriginalPoint.Latitude
			mp.MatchedLon = gp.Lon
			mp.MatchedLat = gp.Lat
			mp.RoadId = -1
			t.Mps = append(t.Mps, mp)
		}
	}

	mp := mmap.GetMatchingPoint()
	mp.OriginalLon = p.Lon
	mp.OriginalLat = p.Lat
	mp.RoadId = p.RoadID

	if p.Ttype == common.DISCRETE {
		mp.MatchedLon = p.OriginalPoint.Longitude
		mp.MatchedLat = p.OriginalPoint.Latitude
		mp.IsBad = 1
	} else {
		mp.MatchedLon = p.Lon
		mp.MatchedLat = p.Lat
		mp.IsBad = 0
	}

	t.Mps = append(t.Mps, mp)
}

func CopyTrack(t *proto_struct.Track) *proto_struct.Track {
	newTrack := &proto_struct.Track{}
	newTrack.Vin = t.Vin
	newTrack.Tid = t.Tid
	newTrack.Probability = t.Probability

	newTrack.Mps = append(newTrack.Mps, t.Mps...)

	return newTrack
}

func DeepCopyTrack(t *proto_struct.Track) *proto_struct.Track {
	newTrack := common.GetTrack()
	newTrack.Vin = t.Vin
	newTrack.Tid = t.Tid
	newTrack.Probability = t.Probability
	newTrack.Mps = make([]*proto_struct.MatchingPoint, 0, len(t.Mps))

	for _, mp := range t.Mps {
		newMp := common.GetMatchingPoint()
		newMp.OriginalLon = mp.OriginalLon
		newMp.OriginalLat = mp.OriginalLat
		newMp.MatchedLon = mp.MatchedLon
		newMp.MatchedLat = mp.MatchedLat
		newMp.RoadId = mp.RoadId
		newMp.IsBad = mp.IsBad
		newTrack.Mps = append(newTrack.Mps, newMp)
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
							common.DebugLog("not exist")
						}
					}
				}
				if startNode == nil {
					//无法连通说明此候选点应当排除
					common.DebugLog("not exist2")
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
			newT := CopyTrack(maxTrack)
			AppendTrack(newT, n, maxPath, mmap)
			newT.Probability = maxScore
			res[n] = newT
		}
	}
	return res
}
