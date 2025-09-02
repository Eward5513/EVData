package main

import (
	"EVdata/common"
	"EVdata/server"
	"time"
)

func main() {

	common.CreatLogFile("server.log")
	defer common.CloseLogFile()
	//common.SetLogLevel(common.DEBUG)

	//限制时区
	time.Local = time.UTC

	worker := server.NewServerWorker()

	worker.Init()

	worker.Start()

	worker.Close()
}
