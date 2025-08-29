package main

import (
	"EVdata/CSV"
	"EVdata/common"
	"EVdata/pgpg"
	"EVdata/proto_tools"
	"log"
	"os"
	"path/filepath"
	"strings"
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

	CreateNewRawData()

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

	//ReadSourceFile(ch)

	//if err := common.DeleteEmptyDirs(common.REFINED_RAW_DATA_DIR_PATH); err != nil {
	//	log.Fatal("error deleting empty dirs:" + err.Error())
	//}
}

func CreateNewRawData() {
	if err := common.ClearFolder(common.REFINED_RAW_DATA_DIR_PATH); err != nil {
		log.Fatalln("Failed to clear folder:", err)
	}

	if err := common.ClearFolder(common.POINT_DATA_DIR_PATH); err != nil {
		log.Fatalln("Failed to clear folder:", err)
	}

	csvFileName := filepath.Join(common.POINT_DATA_DIR_PATH, "points.csv")
	var wFile *os.File
	var err error
	if wFile, err = os.OpenFile(csvFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0744); err != nil {
		log.Fatalln("Failed to create file:", err)
	}
	writer := CSV.NewPointWriter(wFile)
	writerCh := make(chan []*common.RawPoint, 1000)
	wg := &sync.WaitGroup{}

	go func() {
		wg.Add(1)
		for {
			rps := <-writerCh
			if rps == nil {
				wg.Done()
				return
			}
			for _, rp := range rps {
				tp := proto_tools.ConvertRawPointToTrackPoint(rp)
				writer.Write(tp)
			}
		}
	}()

	folderEntries, err := os.ReadDir(common.RAW_DATA_DIR_PATH)

	if err != nil {
		log.Fatalln("Failed to read dir:", err)
	}

	for _, fe1 := range folderEntries {
		if fe1.IsDir() == false {
			continue
		}
		fp := filepath.Join(common.RAW_DATA_DIR_PATH, fe1.Name())
		fe2, err := os.ReadDir(fp)
		if err != nil || len(fe2) != 1 || strings.HasSuffix(fe2[0].Name(), ".parquet") == false {
			log.Fatalln("Failed to read dir:", err)
		}
		rowPoints := pgpg.ReadPointFromParquet(filepath.Join(fp, fe2[0].Name()))
		writerCh <- rowPoints
	}
	wg.Wait()
}
