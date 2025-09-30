package main

import (
	"EVdata/common"
	"EVdata/mapmatching"
	"flag"
	"log"
	"net/http"
	"time"
)

func main() {

	begin := time.Now()

	var profMode bool
	flag.BoolVar(&profMode, "p", false, "pprof mode")
	flag.Float64Var(&mapmatching.DistanceOffset, "d", 40, "Distance offset")
	flag.IntVar(&mapmatching.WorkerCount, "wc", 1000, "worker count")
	flag.Parse()

	common.CreatLogFile("mapmatching.log")
	defer common.CloseLogFile()

	if profMode {
		//common.SetLogLevel(common.DEBUG)
		go func() {
			log.Println("start pprof tool")
			//http://localhost:6060/debug/pprof/
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	common.SetLogLevel(common.INFO)

	t := time.Now()
	mapmatching.BuildGraph("shanghai_new.json")
	common.InfoLog("time for building graph: ", time.Since(t))
	t = time.Now()

	mapmatching.BuildIndex()
	common.InfoLog("time for building index: ", time.Since(t))
	t = time.Now()

	mapmatching.PreComputing()
	common.InfoLog("time for precomputing: ", time.Since(t))

	mapmatching.ProcessDataLoop()

	log.Println("Total Execution time: ", time.Now().Sub(begin))
}
