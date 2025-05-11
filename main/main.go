package main

import (
	"EVdata/common"
	"log"
	"os"
	"sync"
	"time"
)

const (
	UPPER_BUFFER_SIZE  = 1e6
	WORKER_BUFFER_SIZE = 5000
)

func main() {
	//cpuProfile, err := os.OpenFile("cpu.profile", os.O_CREATE|os.O_RDWR, 0644)
	//if err != nil {
	//	log.Println("Cannot open cpu profile: " + err.Error())
	//}
	//defer cpuProfile.Close()
	//memProfile, err := os.OpenFile("mem.profile", os.O_CREATE|os.O_RDWR, 0644)
	//if err != nil {
	//	log.Println("Cannot open memory profile: " + err.Error())
	//}
	//defer memProfile.Close()
	//
	//if err := pprof.StartCPUProfile(cpuProfile); err != nil {
	//	log.Println("Cannot start CPU profile: " + err.Error())
	//}
	//defer pprof.StopCPUProfile()
	//
	//if err := pprof.WriteHeapProfile(memProfile); err != nil {
	//	log.Println("Cannot start memory profile: " + err.Error())
	//}

	begin := time.Now()

	defer HandlePanic()

	common.CreatLogFile("dev.log")
	defer common.CloseLogFile()

	time.Local = time.UTC

	CreateRefinedRawData()

	common.InfoLog("Execution time: ", time.Now().Sub(begin))
}

func HandlePanic() {
	if err := recover(); err != nil {
		log.Println("panic:", err)
	}
}

func CreateRefinedRawData() {
	//if err := common.CreateDirs(common.REFINED_RAW_DATA_DIR_PATH); err != nil {
	//	log.Fatal("error creating dirs:" + err.Error())
	//}

	if err := os.RemoveAll(common.REFINED_RAW_DATA_DIR_PATH); err != nil {
		log.Fatal("Error removing target directory", err.Error())
	}
	if err := os.Mkdir(common.REFINED_RAW_DATA_DIR_PATH, os.ModeDir); err != nil {
		log.Fatal("Error when creating target dir:" + err.Error())
	}

	wg := sync.WaitGroup{}
	ch := make(chan *common.RawPoint, UPPER_BUFFER_SIZE)

	//go CreateCSVManager(&wg, ch)
	go CreateParquetManager(&wg, ch)

	ReadSourceFile(ch)

	close(ch)

	wg.Wait()

	//if err := common.DeleteEmptyDirs(common.REFINED_RAW_DATA_DIR_PATH); err != nil {
	//	log.Fatal("error deleting empty dirs:" + err.Error())
	//}
}
