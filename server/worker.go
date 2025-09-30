package server

import (
	"EVdata/CSV"
	"EVdata/common"
	"EVdata/mapmatching"
	"EVdata/proto_struct"
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PointRequest struct {
	CurrentTime  string  `json:"currentTime"`
	Vin          []int   `json:"vin"`
	MaxLatitude  float64 `json:"maxLatitude"`
	MaxLongitude float64 `json:"maxLongitude"`
	MinLatitude  float64 `json:"minLatitude"`
	MinLongitude float64 `json:"minLongitude"`
	Milliseconds int64
}

// 响应体结构
type PointResponse struct {
	Data []proto_struct.RawPoint `json:"points"`
}

type channelRequest struct {
	startTime *time.Time
	endTime   *time.Time
	vin       int
}

type ServerWorker struct {
	OrderCh   []chan *channelRequest
	ReceiveCh chan []*proto_struct.RawPoint
	wg        *sync.WaitGroup
	data      [][]*proto_struct.RawPoint
	graph     map[int64]*common.GraphNode
}

func NewServerWorker() *ServerWorker {
	sw := &ServerWorker{}
	sw.OrderCh = make([]chan *channelRequest, common.SERVER_WORKER_COUNT)
	sw.ReceiveCh = make(chan []*proto_struct.RawPoint)
	sw.data = make([][]*proto_struct.RawPoint, common.VEHICLE_COUNT+1)
	sw.wg = &sync.WaitGroup{}
	for i := 0; i < common.SERVER_WORKER_COUNT; i++ {
		sw.OrderCh[i] = make(chan *channelRequest, 256)
	}
	return sw
}

func (sw *ServerWorker) Init() {
	//sw.data = pgpg.ReadPointFromParquet(common.RAW_POINT_PARQUET_PATH)

	// 建图操作
	mapmatching.BuildGraph("shanghai_new.json")
	sw.graph = mapmatching.Graph

	for i := 0; i < common.SERVER_WORKER_COUNT; i++ {
		go sw.work(sw.OrderCh[i], i)
	}
}

func (sw *ServerWorker) Start() {
	http.HandleFunc("/api/point", sw.PointHandler)
	http.HandleFunc("/api/track", sw.TrackHandler)
	http.HandleFunc("/api/generateTrack", sw.GenerateTrackHandler)

	// 启动 Golang 后端
	log.Println("Golang backend is running on http://127.0.0.1:3000")
	if err := http.ListenAndServe("localhost:3000", nil); err != nil {
		log.Println(err.Error())
	}
}

func (sw *ServerWorker) Close() {
	for i := 0; i < common.SERVER_WORKER_COUNT; i++ {
		close(sw.OrderCh[i])
	}
	sw.wg.Wait()
}

func (sw *ServerWorker) PointHandler(w http.ResponseWriter, r *http.Request) {
	// 检查请求方法是否为 POST
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Expose-Headers", "Access-Control-Allow-Origin,Content-Type")

	//解析请求体
	var request PointRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	t, err := time.Parse("2006-01-02T15:04:05Z", request.CurrentTime)
	if err != nil {
		log.Println(err)
	}

	log.Println("Receive data", request.CurrentTime, request.Vin, request.MaxLatitude, request.MinLatitude, request.MaxLongitude, request.MinLongitude)

	targetCount := common.VEHICLE_COUNT
	if len(request.Vin) != 0 {
		targetCount = len(request.Vin)
	}

	go func() {
		// 读取所有数据
		if len(request.Vin) == 0 {
			for i := 1; i <= common.VEHICLE_COUNT; i++ {
				sw.OrderCh[i%common.SERVER_WORKER_COUNT] <- &channelRequest{
					startTime: &t,
					endTime:   &t,
					vin:       i,
				}
			}
		} else {
			for i := 0; i < len(request.Vin); i++ {
				sw.OrderCh[i] <- &channelRequest{
					startTime: &t,
					endTime:   &t,
					vin:       request.Vin[i],
				}
			}
		}
	}()

	response := PointResponse{Data: make([]proto_struct.RawPoint, 0, targetCount)}
	var cnt int
	for ps := range sw.ReceiveCh {
		for _, p := range ps {
			//log.Println("Append data", p, request, response)
			if len(request.Vin) != 0 ||
				(p.Latitude <= request.MaxLatitude &&
					p.Latitude >= request.MinLatitude &&
					p.Longitude <= request.MaxLongitude &&
					p.Longitude >= request.MinLongitude) {
				response.Data = append(response.Data, proto_struct.RawPoint{
					Latitude:  p.Latitude,
					Longitude: p.Longitude,
					Vin:       p.Vin,
					Time:      p.Time,
				})
			}

		}
		cnt++
		//log.Println(cnt)
		if cnt >= targetCount {
			log.Println("Get all data", cnt, len(response.Data))
			break
		}
	}

	//返回响应
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

type GenerateTrackRequest struct {
	Vin       string `json:"vin"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type GenerateTrackResponse struct {
	Data []*proto_struct.TrackPoint `json:"track_points"`
}

func (sw *ServerWorker) GenerateTrackHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("GenerateTrackHandler")

	// 检查请求方法是否为 POST
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Expose-Headers", "Access-Control-Allow-Origin,Content-Type")

	// 解析请求体
	var request GenerateTrackRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	fp := filepath.Join(common.MATCHED_RAW_POINT_CSV_DIR, request.Vin+".csv")
	if _, err := os.Stat(fp); err != nil {
		http.Error(w, "Invalid Vin or Tid", http.StatusBadRequest)
		return
	}

	// 读取轨迹数据
	vinInt, _ := strconv.Atoi(request.Vin)
	mps := CSV.ReadTrackPointFromCSV(vinInt)

	startTimeInt := common.ParseTimeToInt(request.StartTime)
	endTimeInt := common.ParseTimeToInt(request.EndTime)

	res := make([]*proto_struct.TrackPoint, 0, len(mps))
	for _, mp := range mps {
		if mp.TimeInt >= startTimeInt && mp.TimeInt <= endTimeInt {
			res = append(res, mp)
		}
	}

	// 构建响应
	response := GenerateTrackResponse{
		Data: res,
	}

	// 返回响应
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		common.ErrorLog("Error writing response", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

type TrackRequest struct {
	Vin int `json:"vin"`
	Tid int `json:"tid"`
}

// 轨迹查询响应结构
type TrackResponse struct {
	Data []*proto_struct.Track `json:"tracks"`
}

// 新的轨迹点响应结构
type TrackPointResponse struct {
	Data []TrackPointData `json:"points"`
}

type TrackPointData struct {
	Vin       int     `json:"vin"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Timestamp string  `json:"timestamp"`
}

func (sw *ServerWorker) TrackHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("TrackHandler")

	// 检查请求方法是否为 POST
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Expose-Headers", "Access-Control-Allow-Origin,Content-Type")

	// 解析请求体
	var request TrackRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 读取三个文件的数据
	queryData, err := sw.readQueryFile()
	if err != nil {
		http.Error(w, "Failed to read query.txt: "+err.Error(), http.StatusInternalServerError)
		return
	}

	routeData, err := sw.readRouteFile()
	if err != nil {
		http.Error(w, "Failed to read route.txt: "+err.Error(), http.StatusInternalServerError)
		return
	}

	timeData, err := sw.readTimeFile()
	if err != nil {
		http.Error(w, "Failed to read time.txt: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 直接使用tid作为行索引
	common.InfoLog(len(queryData), len(routeData), len(timeData))
	if request.Tid >= len(queryData) || request.Tid >= len(routeData) || request.Tid >= len(timeData) {
		http.Error(w, "Invalid tid: index out of range", http.StatusBadRequest)
		return
	}

	targetQuery := strings.Fields(queryData[request.Tid])
	targetRoute := strings.Fields(routeData[request.Tid])
	targetTime := strings.Fields(timeData[request.Tid])

	// 解析数据并构建响应点列表
	var points []TrackPointData

	if len(targetRoute) >= 2 && len(targetTime) >= 2 {
		// 解析route数据: 节点数量 node1 bool1 bool2 bool3 bool4 node2 ...
		nodeCount, _ := strconv.Atoi(targetRoute[0])

		// 提取节点ID和时间戳
		nodeIndex := 1
		timeIndex := 1

		for i := 0; i < nodeCount && nodeIndex < len(targetRoute); i++ {
			nodeId, err := strconv.ParseInt(targetRoute[nodeIndex], 10, 64)
			if err != nil {
				nodeIndex += 5 // 跳过当前节点和4个bool值
				continue
			}

			// 从图中获取经纬度
			if node, exists := sw.graph[nodeId]; exists {
				var timestampMillis int64
				if i == 0 {
					// 第一个点使用query中的开始时间
					if len(targetQuery) >= 4 {
						timestampMillis, _ = strconv.ParseInt(targetQuery[2], 10, 64)
					}
				} else if timeIndex < len(targetTime) {
					// 其他点使用time文件中的时间
					timestampMillis, _ = strconv.ParseInt(targetTime[timeIndex], 10, 64)
					timeIndex++
				}

				// 将毫秒级时间戳转换为 hh:mm:ss 格式（使用 UTC）
				timestamp := time.UnixMilli(timestampMillis).UTC().Format("15:04:05")

				points = append(points, TrackPointData{
					Vin:       request.Vin,
					Longitude: node.Lon,
					Latitude:  node.Lat,
					Timestamp: timestamp,
				})
			}

			nodeIndex += 5 // 移动到下一个节点（跳过4个bool值）
		}
	}

	// 构建响应
	response := TrackPointResponse{
		Data: points,
	}

	// 返回响应
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		common.ErrorLog("Error writing response", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (sw *ServerWorker) work(ch chan *channelRequest, cnt int) {
	var points []*proto_struct.RawPoint

	sw.wg.Add(1)
	defer sw.wg.Done()

	targetPoints := make([]*proto_struct.RawPoint, 0)

	log.Println("worker start", cnt)

	for {
		ts := <-ch
		if ts == nil {
			return
		}
		targetPoints = targetPoints[:0]
		startTimeInt := common.ParseTimeToInt(ts.startTime.Format("15:04:05"))
		endTimeInt := common.ParseTimeToInt(ts.endTime.Format("15:04:05"))
		points = sw.data[ts.vin]
		//log.Println("parse time:", startTimeInt, endTimeInt)

		if len(points) != 0 {
			//寻找最后一个比目标时间小的点
			left, right := 0, len(points)-1
			var mid int
			res := 0
			for left <= right {
				mid = (left + right) >> 1
				if points[mid].TimeInt <= startTimeInt {
					res = mid
					left = mid + 1
				} else {
					right = mid - 1
				}
			}
			//如果只需要一个点
			if startTimeInt == endTimeInt {
				if common.Abs(points[res].TimeInt, startTimeInt) <= 30 {
					targetPoints = append(targetPoints, points[res])
				}
			} else {
				targetPtr := -1
				for i := res; i < len(points) && points[i].TimeInt <= endTimeInt; i++ {
					//if vin == 1 {
					//	log.Println(points[i].CollectionTime)
					//}
					//取多个点时需要去重
					if points[i].TimeInt >= startTimeInt && (targetPtr == -1 || (targetPoints[targetPtr].Longitude != points[i].Longitude && targetPoints[targetPtr].Latitude != points[i].Latitude)) {
						points[i].Vin = int32(ts.vin)
						targetPoints = append(targetPoints, points[i])
						targetPtr++
					}
				}
			}
		}
		//...
		sw.ReceiveCh <- targetPoints
	}
}

// 读取query.txt文件
func (sw *ServerWorker) readQueryFile() ([]string, error) {
	file, err := os.Open(filepath.Join(common.TRACK_DATA_DIR_PATH, "query.txt"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// 读取route.txt文件
func (sw *ServerWorker) readRouteFile() ([]string, error) {
	file, err := os.Open(filepath.Join(common.TRACK_DATA_DIR_PATH, "route.txt"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// 读取time.txt文件
func (sw *ServerWorker) readTimeFile() ([]string, error) {
	file, err := os.Open(filepath.Join(common.TRACK_DATA_DIR_PATH, "time.txt"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
