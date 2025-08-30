package main

import (
	"EVdata/CSV"
	"EVdata/common"
	"EVdata/pgpg"
	"EVdata/proto_tools"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CreateCSVWorker 创建所需协程及通道
//func CreateCSVWorker(wg *sync.WaitGroup) map[int]chan *common.RawPoint {
//	years := []int{2021, 2022}
//	month, day := 13, 32
//	register := make(map[int]chan *common.RawPoint)
//	//年
//	for y := 0; y < len(years); y++ {
//		//月
//		for m := 1; m < month; m++ {
//			//日
//			for d := 1; d < day; d++ {
//				t := years[y]*1e4 + m*1e2 + d
//				ch := make(chan *common.RawPoint, WORKER_BUFFER_SIZE)
//				register[t] = ch
//				//log.Println("Register ", t)
//				dirPath := filepath.Join(common.REFINED_RAW_DATA_DIR_PATH, fmt.Sprint(years[y]), fmt.Sprint(m), fmt.Sprint(d))
//				go WriteDataToCSV(dirPath, ch, wg)
//			}
//		}
//	}
//	return register
//}

//func CreateCSVManager(wg *sync.WaitGroup, ch chan *common.RawPoint) {
//	wg.Add(1)
//	log.Println("Create Manager")
//	defer wg.Done()
//	register := CreateCSVWorker(wg)
//	for {
//		ms := <-ch
//		//log.Println("Upper channel:", len(ch))
//		if ms != nil {
//			//log.Println("manager receive data", ms)
//			if _, ok := register[ms.TimeStamp]; ok == false {
//				log.Println("Unknown time", ms.TimeStamp, ms.CollectionTime)
//			} else {
//				register[ms.TimeStamp] <- ms
//			}
//		} else {
//			//通道关闭，关闭所有worker
//			log.Println("Close Manager")
//			for _, v := range register {
//				close(v)
//			}
//			return
//		}
//	}
//}

func ReadSourceFile(ch chan *common.RawPoint) {
	sm1, sm2 := common.NewSummary(), common.NewSummary()
	files, err := os.ReadDir(common.RAW_DATA_DIR_PATH)
	if err != nil {
		log.Fatal("Unable to read source path", err.Error())
	}
	for i, file := range files {
		//仅处理文件夹
		if file.IsDir() {
			//取出vin
			//var vin string
			//_, err = fmt.Sscanf(file.Name(), "vin=%s", &vin)
			//if err != nil {
			//	log.Println("Unable to get vin from", file.Name(), err.Error())
			//}
			vin := fmt.Sprintf("%d", i-1)

			//读取parquet文件
			dirName := filepath.Join(common.RAW_DATA_DIR_PATH, file.Name())
			parquetFiles, err := os.ReadDir(dirName)
			if err != nil {
				log.Println("Unable to read parquet files in", file.Name(), err.Error())
			}
			for _, parquetFile := range parquetFiles {
				if strings.HasSuffix(parquetFile.Name(), "parquet") {
					//读文件
					rows := pgpg.ReadPointFromParquet(filepath.Join(dirName, parquetFile.Name()))
					for _, row := range rows {
						//将数据发送给对应worker
						row.Vin = vin
						row.CollectionTime += common.TimeOffset
						t := time.UnixMilli(row.CollectionTime)
						row.TimeStamp = t.Year()*1e4 + int(t.Month())*1e2 + t.Day()
						sm1.Operate(row)
						if row.IsValid() {
							sm2.Operate(row)
							ch <- row
						}
					}
				}
			}
			common.InfoLog("Finish", vin, "/", common.PARQUET_COUNT)
		}
	}
	sm1.PrintData()
	sm2.PrintData()
}

func WriteDataToCSV(dirPath string, ch chan *common.RawPoint, wg *sync.WaitGroup) {
	wg.Add(1)
	log.Println("worker starts:", dirPath)
	defer wg.Done()

	var vin string
	var csvFile *os.File
	var csvWriter *CSV.MatchingPointWriter
	for {
		ms := <-ch
		//log.Println("worker channel", len(ch))
		if ms == nil {
			if csvWriter != nil {
				csvWriter.Close()
			}
			common.DebugLog("worker closed by channel", dirPath)
			return
		}
		//log.Println("receive message:", ms)
		if vin != ms.Vin {
			if len(vin) != 0 && csvWriter != nil {
				csvWriter.Close()
				//log.Println("write to file", csvWriter.Name())
			}
			n := filepath.Join(dirPath, ms.Vin+".csv")
			csvFile, _ = os.Create(n)
			csvWriter = CSV.NewMatchingPointWriter(csvFile)
			//log.Println("create new file:", n)
		}
		//if err := csvWriter.Write(ms); err != nil {
		//	log.Println("Error writing csv file:", err.Error())
		//}
		vin = ms.Vin
		//log.Println(dirPath, " receives message", ms)
	}
}

func CreateParquetManager(wg *sync.WaitGroup, ch chan *common.RawPoint) {
	wg.Add(1)
	common.InfoLog("Create Manager")
	defer wg.Done()
	register := CreateParquetWriter(wg)
	var currentVin int
	var currentCh chan *common.RawPoint
	closedChannel := make([]int, common.PARQUET_COUNT+1)
	for {
		ms := <-ch
		//log.Println("Upper channel:", len(ch))
		if ms != nil {
			//log.Println("manager receive data", ms)
			v, err := strconv.Atoi(ms.Vin)
			if err != nil {
				common.ErrorLog("Parquet worker error:", err.Error())
			}
			if _, ok := register[v]; ok == false {
				common.ErrorLog("Unknown vin", ms.Vin)
			} else {
				if currentVin != v {
					if currentCh != nil {
						//common.InfoLog("close channel", currentVin)
						close(currentCh)
						closedChannel[currentVin] = 1
					}
					currentVin = v
					currentCh = register[currentVin]
				}
				currentCh <- ms
			}
		} else {
			//通道关闭，关闭manager
			if currentCh != nil {
				//common.InfoLog("close channel", currentVin)
				close(currentCh)
				closedChannel[currentVin] = 1
			}
			for i := 1; i <= common.PARQUET_COUNT; i++ {
				if closedChannel[i] == 0 {
					//common.InfoLog("close channel", i)
					close(register[i])
				}
			}
			common.InfoLog("Close Parquet Manager")
			return
		}
	}
}

func CreateParquetWriter(wg *sync.WaitGroup) map[int]chan *common.RawPoint {
	register := make(map[int]chan *common.RawPoint)
	for i := 1; i <= common.PARQUET_COUNT; i++ {
		ch := make(chan *common.RawPoint, WORKER_BUFFER_SIZE)
		register[i] = ch
		//log.Println("Register ", t)
		wg.Add(1)
		fPath := filepath.Join(common.REFINED_RAW_DATA_DIR_PATH, fmt.Sprint(i)+".parquet")
		go ParquetWriter(fPath, ch, wg)

	}
	return register
}

func ParquetWriter(fPath string, ch chan *common.RawPoint, wg *sync.WaitGroup) {
	log.Println("worker starts:", fPath)
	defer wg.Done()

	pqWriter := pgpg.NewPQWriter(fPath)
	for {
		ms := <-ch
		//log.Println("worker channel", len(ch))
		if ms == nil {
			if pqWriter != nil {
				pqWriter.Close()
			}
			common.DebugLog("worker closed by channel", fPath)
			return
		}
		tp := proto_tools.ConvertRawPointToTrackPoint(ms)
		pqWriter.Write(tp)
	}
}
