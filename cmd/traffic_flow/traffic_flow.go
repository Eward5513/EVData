package main

import (
	"EVdata/common"
	"EVdata/proto_struct"
	"EVdata/traffic_flow"
	"time"
)

var (
	readerChannelSize = 100
	writerChannelSize = 100
)

func main() {
	begin := time.Now()

	readerCh := make(chan []*proto_struct.TrackPoint, readerChannelSize)
	writerCh := make(chan []*common.TrafficFlow, writerChannelSize)

	go traffic_flow.StartReader(readerCh)
	go traffic_flow.StartWorker(readerCh, writerCh)

	traffic_flow.StartWriter(writerCh)

	//traffic_flow.GenerateNetwork()

	common.InfoLog("Total Execution time: ", time.Now().Sub(begin))
}
