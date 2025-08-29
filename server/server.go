package main

import (
	"EVdata/common"
	"time"
)

func main() {

	common.CreatLogFile("server.log")
	defer common.CloseLogFile()
	//common.SetLogLevel(common.DEBUG)

	//限制时区
	time.Local = time.UTC

	worker := NewServerWorker()

	worker.Init()

	worker.Start()

	worker.Close()
}
