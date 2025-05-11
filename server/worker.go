package main

import (
	"EVdata/CSV"
	"EVdata/common"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
type TrackResponse struct {
	Data []common.TrackPoint `json:"track"`
}

type TrackRequest struct {
	StartTime    string  `json:"startTime"`
	EndTime      string  `json:"endTime"`
	Vin          []int   `json:"vin"`
	MaxLatitude  float64 `json:"maxLatitude"`
	MaxLongitude float64 `json:"maxLongitude"`
	MinLatitude  float64 `json:"minLatitude"`
	MinLongitude float64 `json:"minLongitude"`
	Milliseconds int64
}

// 响应体结构
type PointResponse struct {
	Data []common.TrackPoint `json:"points"`
}

type channelRequest struct {
	startTime *time.Time
	endTime   *time.Time
	vin       []int
}

type ServerWorker struct {
	OrderCh   []chan *channelRequest
	ReceiveCh chan []*common.TrackPoint
	wg        *sync.WaitGroup
}

func NewServerWorker() *ServerWorker {
	sw := &ServerWorker{}
	sw.OrderCh = make([]chan *channelRequest, common.VehicleCount+1)
	sw.ReceiveCh = make(chan []*common.TrackPoint)
	sw.wg = &sync.WaitGroup{}
	for i := 1; i <= common.VehicleCount; i++ {
		sw.OrderCh[i] = make(chan *channelRequest, 256)
	}
	return sw
}

func (sw *ServerWorker) Init() {
	for i := 1; i <= common.VehicleCount; i++ {
		go sw.work(i, sw.OrderCh[i])
	}
}

func (sw *ServerWorker) Start() {
	http.HandleFunc("/api/point", sw.PointHandler)
	http.HandleFunc("/api/track", sw.TrackHandler)
	http.HandleFunc("/api/queryTrack", sw.TrackQueryHandler)
	http.HandleFunc("/api/dates", sw.GetDatesHandler)

	// 启动 Golang 后端
	log.Println("Golang backend is running on http://127.0.0.1:3000")
	if err := http.ListenAndServe("localhost:3000", nil); err != nil {
		log.Println(err.Error())
	}
}

func (sw *ServerWorker) Close() {
	for i := 0; i < common.VehicleCount; i++ {
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

	cr := &channelRequest{
		startTime: &t,
		endTime:   &t,
		vin:       request.Vin,
	}

	log.Println("Receive data", request.CurrentTime, t.UnixMilli(), request.Vin, request.MaxLatitude, request.MinLatitude, request.MaxLongitude, request.MinLongitude)

	targetVin := make(map[int]bool)
	for _, v := range request.Vin {
		targetVin[v] = true
	}

	targetCount := common.VehicleCount
	if len(targetVin) != 0 {
		targetCount = len(targetVin)
	}

	// 从1开始有效
	for i := 1; i < len(sw.OrderCh); i++ {
		if _, exist := targetVin[i]; len(targetVin) == 0 || exist == true {
			//log.Println("send data", i)
			sw.OrderCh[i] <- cr
		}
	}
	response := PointResponse{Data: make([]common.TrackPoint, 0, targetCount)}
	var cnt int
	for ps := range sw.ReceiveCh {
		for _, p := range ps {
			//log.Println("Append data", p, request, response)
			if len(request.Vin) != 0 ||
				(p.Latitude <= request.MaxLatitude &&
					p.Latitude >= request.MinLatitude &&
					p.Longitude <= request.MaxLongitude &&
					p.Longitude >= request.MinLongitude) {
				response.Data = append(response.Data, common.TrackPoint{
					Latitude:  p.Latitude,
					Longitude: p.Longitude,
					Vin:       p.Vin,
				})
			}
		}
		cnt++
		if cnt >= targetCount {
			log.Println("Get all data", cnt, len(response.Data))
			break
		}
	}

	//返回响应
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func (sw *ServerWorker) TrackHandler(w http.ResponseWriter, r *http.Request) {
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
	var request TrackRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	st, err := time.Parse("2006-01-02T15:04:05Z", request.StartTime)
	if err != nil {
		log.Println(err)
	}
	et, err := time.Parse("2006-01-02T15:04:05Z", request.EndTime)
	if err != nil {
		log.Println(err)
	}
	cr := &channelRequest{
		startTime: &st,
		endTime:   &et,
		vin:       request.Vin,
	}

	log.Println("Receive data", cr.startTime, cr.endTime, cr.vin, request.MaxLatitude, request.MinLatitude, request.MaxLongitude, request.MinLongitude)

	targetVin := make(map[int]bool)
	for _, v := range request.Vin {
		targetVin[v] = true
	}

	targetCount := common.VehicleCount
	if len(targetVin) != 0 {
		targetCount = len(targetVin)
	}

	// 从1开始有效
	for i := 1; i < len(sw.OrderCh); i++ {
		if _, exist := targetVin[i]; len(targetVin) == 0 || exist == true {
			//log.Println("send data", i)
			sw.OrderCh[i] <- cr
		}
	}
	response := TrackResponse{Data: make([]common.TrackPoint, 0, common.VehicleCount)}
	var cnt int
	for ps := range sw.ReceiveCh {
		for _, p := range ps {
			log.Println("Append data", p.Vin, p.Timestamp)
			if len(request.Vin) != 0 ||
				(p.Latitude <= request.MaxLatitude &&
					p.Latitude >= request.MinLatitude &&
					p.Longitude <= request.MaxLongitude &&
					p.Longitude >= request.MinLongitude) {
				response.Data = append(response.Data, common.TrackPoint{
					Latitude:  p.Latitude,
					Longitude: p.Longitude,
					Vin:       p.Vin,
					Timestamp: p.Timestamp,
					Date:      p.Date,
					Hour:      p.Hour,
					Speed:     p.Speed,
				})
			}
		}
		cnt++
		if cnt >= targetCount {
			log.Println("Get all data", cnt)
			break
		}
	}

	//返回响应
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

type TrackQueryRequest struct {
	Vin         int    `json:"vin"`
	Tid         int    `json:"tid"`
	CurrentTime string `json:"currentTime"`
}

// 轨迹查询响应结构
type TrackQueryResponse struct {
	Data []*common.Track `json:"tracks"`
}

func (sw *ServerWorker) TrackQueryHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("TrackQueryHandler")

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
	var request TrackQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 构建CSV文件路径
	t, err := time.Parse("2006-01-02T15:04:05Z", request.CurrentTime)
	if err != nil {
		common.ErrorLog("Error parsing time", err)
		http.Error(w, "Invalid current time", http.StatusBadRequest)
	}
	csvPath := filepath.Join(common.TRACK_DATA_DIR_PATH, fmt.Sprint(t.Year()), fmt.Sprintf("%d", t.Month()), fmt.Sprint(t.Day()), fmt.Sprintf("%d_%d.csv", request.Vin, request.Tid))

	// 读取轨迹数据
	tr, err := CSV.ReadTrackFromCSV(csvPath)
	if err != nil {
		common.ErrorLog("Error when reading track: ", err)
		http.Error(w, "Failed to read track data", http.StatusInternalServerError)
		return
	}

	// 构建响应
	response := TrackQueryResponse{
		Data: []*common.Track{tr},
	}

	// 返回响应
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		common.ErrorLog("Error writing response", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (sw *ServerWorker) work(id int, ch chan *channelRequest) {
	var curFile string
	var points []*common.TrackPoint

	vin := id

	sw.wg.Add(1)
	defer sw.wg.Done()

	log.Println("worker start", vin)

	targetPoints := make([]*common.TrackPoint, 0)

	for {
		ts := <-ch
		if ts == nil {
			return
		}
		//log.Println("Get data", ts)
		targetPoints = targetPoints[:0]
		startMillSeconds := ts.startTime.UnixMilli()
		endMillSeconds := ts.endTime.UnixMilli()
		//common.DebugLog("time", startMillSeconds, endMillSeconds)

		targetFile := filepath.Join(common.POINT_DATA_DIR_PATH, fmt.Sprintf("%d\\%d\\%d\\%d.csv", ts.startTime.Year(), ts.startTime.Month(), ts.startTime.Day(), vin))
		//如果需要读取新文件
		if targetFile != curFile {
			//目标文件必须存在
			if _, err := os.Stat(targetFile); os.IsNotExist(err) == false {
				log.Println("read file", targetFile)
				curFile = targetFile
				points = CSV.ReadPointFromCSV(curFile)
				log.Println(len(points))
			} else {
				log.Println("error", err)
			}
		}
		if len(points) != 0 {
			//寻找最后一个比目标时间小的点
			left, right := 0, len(points)-1
			var mid int
			res := 0
			for left <= right {
				mid = (left + right) >> 1
				if points[mid].CollectionTime <= startMillSeconds {
					res = mid
					left = mid + 1
				} else {
					right = mid - 1
				}
			}
			//如果只需要一个点
			//if vin == 1 {
			//	log.Println(endMillSeconds, startMillSeconds, res, len(points))
			//}
			if endMillSeconds == startMillSeconds {
				if common.Abs(points[res].CollectionTime, startMillSeconds) <= 60000 {
					targetPoints = append(targetPoints, points[res])
				}
			} else {
				targetPtr := -1
				for i := res; i < len(points) && points[i].CollectionTime <= endMillSeconds; i++ {
					//if vin == 1 {
					//	log.Println(points[i].CollectionTime)
					//}
					//取多个点时需要去重
					if points[i].CollectionTime >= startMillSeconds && (targetPtr == -1 || (targetPoints[targetPtr].Longitude != points[i].Longitude && targetPoints[targetPtr].Latitude != points[i].Latitude)) {
						points[i].Vin = vin
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
func (sw *ServerWorker) GetDatesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 获取日期数据
	dateData, err := getAllDates()
	if err != nil {
		log.Printf("Error getting dates: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 返回JSON数据
	if err := json.NewEncoder(w).Encode(dateData); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// getAllDates 遍历目录获取所有有效日期
func getAllDates() (map[string]map[string][]int, error) {
	result := make(map[string]map[string][]int)

	// 遍历年份目录
	years, err := os.ReadDir(common.TRACK_DATA_DIR_PATH)
	if err != nil {
		return nil, err
	}

	for _, yearInfo := range years {
		if !yearInfo.IsDir() {
			continue
		}

		yearName := yearInfo.Name()
		yearPath := filepath.Join(common.TRACK_DATA_DIR_PATH, yearName)

		// 为每年创建月份映射
		result[yearName] = make(map[string][]int)

		// 遍历月份目录
		months, err := os.ReadDir(yearPath)
		if err != nil {
			log.Printf("Error reading month directory for year %s: %v", yearName, err)
			continue
		}

		for _, monthInfo := range months {
			if !monthInfo.IsDir() {
				continue
			}

			monthName := monthInfo.Name()
			monthPath := filepath.Join(yearPath, monthName)

			// 遍历日期目录
			days, err := os.ReadDir(monthPath)
			if err != nil {
				log.Printf("Error reading day directory for year %s, month %s: %v", yearName, monthName, err)
				continue
			}

			var daysArray []int
			for _, dayInfo := range days {
				if !dayInfo.IsDir() {
					continue
				}

				dayName := dayInfo.Name()
				dayNum, err := strconv.Atoi(dayName)
				if err != nil {
					log.Printf("Invalid day format: %s", dayName)
					continue
				}

				daysArray = append(daysArray, dayNum)
			}

			// 只有当有效日期存在时才添加到结果中
			if len(daysArray) > 0 {
				result[yearName][monthName] = daysArray
			}
		}

		// 如果该年没有有效月份，则删除该年
		if len(result[yearName]) == 0 {
			delete(result, yearName)
		}
	}

	return result, nil
}
