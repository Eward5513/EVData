package main

import (
	"EVdata/common"
	"EVdata/proto_struct"
	"EVdata/traffic_flow"
)

var (
	readerChannelSize = 100
	writerChannelSize = 100
)

func main() {
	readerCh := make(chan *proto_struct.TrackPoint, readerChannelSize)
	writerCh := make(chan []*common.TrafficFlow, writerChannelSize)

	traffic_flow.StartReader(readerCh)
	traffic_flow.StartWriter(writerCh)
	traffic_flow.StartWorker(readerCh, writerCh)
}
