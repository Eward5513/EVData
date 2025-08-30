package main

import (
	"EVdata/CSV"
	"EVdata/common"
	"EVdata/pgpg"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const READ_BUFFER_SIZE = 5e5
const WRITE_BUFFER_SIZE = 1e3

func main() {
	begin := time.Now()

	common.CreatLogFile("refine_point_data.log")
	defer common.CloseLogFile()

	CreateRefinedPointData()

	common.InfoLog("Execution time: ", time.Now().Sub(begin))
}

func CreateRefinedPointData() {
	if err := common.CreateDirs(common.POINT_DATA_DIR_PATH); err != nil {
		log.Fatal("error creating dirs:" + err.Error())
	}

	wg := sync.WaitGroup{}
	ch := make(chan *common.TrackPoint, READ_BUFFER_SIZE)

	go CreateCSVManager(&wg, ch)

	ReadParquetFile(ch)

	wg.Wait()

	if err := common.DeleteEmptyDirs(common.POINT_DATA_DIR_PATH); err != nil {
		log.Fatal("error deleting empty dirs:" + err.Error())
	}
}

func CreateCSVManager(wg *sync.WaitGroup, ch chan *common.TrackPoint) {
	wg.Add(1)
	common.InfoLog("Create CSV Manager")
	defer wg.Done()
	register := CreateCSVWorker(wg)
	var cnt int
	for {
		ms := <-ch
		//log.Println("Upper channel:", len(ch))
		if ms != nil {
			//log.Println("manager receive data", ms)
			var y, m, d, key int
			//log.Println(ms.Date, len(ms.Date))
			if _, err := fmt.Sscanf(ms.Date, "%d-%d-%d", &y, &m, &d); err != nil {
				common.ErrorLog("error when reading date", err)
			}
			key = y*1e4 + m*1e2 + d

			if wch, ok := register[key]; ok == false {
				log.Println("Unknown time", key, ms.Vin, ms.Date)
			} else {
				cnt++
				wch <- ms
			}
		} else {
			//通道关闭，关闭所有worker
			common.InfoLog("Close Manager")
			common.InfoLog("writing count:", cnt)
			for _, v := range register {
				close(v)
			}
			return
		}
	}
}

// CreateCSVWorker 创建所需协程及通道
func CreateCSVWorker(wg *sync.WaitGroup) map[int]chan *common.TrackPoint {
	years := []int{2021, 2022}
	month, day := 13, 32
	register := make(map[int]chan *common.TrackPoint)
	//年
	for y := 0; y < len(years); y++ {
		//月
		for m := 1; m < month; m++ {
			//日
			for d := 1; d < day; d++ {
				t := years[y]*1e4 + m*1e2 + d
				ch := make(chan *common.TrackPoint, WRITE_BUFFER_SIZE)
				register[t] = ch
				//log.Println("Register ", t)
				dirPath := filepath.Join(common.POINT_DATA_DIR_PATH, fmt.Sprint(years[y]), fmt.Sprint(m), fmt.Sprint(d))
				go WriteDataToCSV(dirPath, ch, wg)
			}
		}
	}
	return register
}

func WriteDataToCSV(dirPath string, ch chan *common.TrackPoint, wg *sync.WaitGroup) {
	wg.Add(1)
	common.InfoLog("worker starts:", dirPath)
	defer wg.Done()

	var csvWriter *CSV.MatchingPointWriter
	for {
		ms := <-ch
		//log.Println("worker channel", len(ch))
		if ms == nil {
			common.InfoLog("worker closed by channel", dirPath)
			return
		}
		f := filepath.Join(dirPath, fmt.Sprintf("%d_%d.csv", ms.Vin, ms.Date))
		fi, err := os.Create(f)
		if err != nil {
			common.ErrorLog(err)
		}
		csvWriter = CSV.NewMatchingPointWriter(fi)
		csvWriter.Write(ms)
		csvWriter.Close()
	}
}

func ReadParquetFile(rch chan *common.TrackPoint) {
	basePath := common.REFINED_RAW_DATA_DIR_PATH
	fs, err := os.ReadDir(basePath)
	if err != nil {
		log.Fatal("Unable to read top dir", err.Error())
	}

	var cnt int
	for i, f := range fs {
		if filepath.Ext(f.Name()) == ".parquet" {
			rows := pgpg.ReadTrackPointFromParquet(filepath.Join(basePath, f.Name()))
			cnt += len(rows)
			for _, row := range rows {
				rch <- row
			}
		}
		common.InfoLog("Finish reading ", i)
	}
	common.InfoLog("reading count:", cnt)
	close(rch)
}
