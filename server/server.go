package main

import (
	"EVdata/common"
	"flag"
	"log"
	"time"
)

func main() {

	common.CreatLogFile("server.log")
	defer common.CloseLogFile()
	//common.SetLogLevel(common.DEBUG)

	flag.IntVar(&common.VehicleCount, "vc", common.PARQUET_COUNT, "vehicle count")
	flag.Parse()

	//限制时区
	time.Local = time.UTC

	worker := NewServerWorker()

	worker.Init()

	worker.Start()

	log.Println("RawPoint Count ", common.VehicleCount)

	worker.Close()
}
