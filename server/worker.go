package main

import (
	"EVdata/CSV"
	"EVdata/common"
	"EVdata/proto_struct"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
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

	for i := 0; i < common.SERVER_WORKER_COUNT; i++ {
		go sw.work(sw.OrderCh[i], i)
	}
}

func (sw *ServerWorker) Start() {
	http.HandleFunc("/api/point", sw.PointHandler)
	http.HandleFunc("/api/track", sw.TrackHandler)

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

type TrackRequest struct {
	Vin         int    `json:"vin"`
	Tid         int    `json:"tid"`
	CurrentTime string `json:"currentTime"`
}

// 轨迹查询响应结构
type TrackResponse struct {
	Data []*proto_struct.Track `json:"tracks"`
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

	// 读取轨迹数据
	mps := CSV.ReadMatchingPointFromCSV(filepath.Join(common.MATCHED_POINT_CSV_DIR, fmt.Sprintf("%d_%d.csv", request.Vin, request.Tid)))

	t := &proto_struct.Track{
		Vin: int32(request.Vin),
		Tid: int32(request.Tid),
		Mps: mps,
	}

	// 构建响应
	response := TrackResponse{
		Data: []*proto_struct.Track{t},
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
