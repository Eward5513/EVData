package CSV

import (
	"EVdata/common"
	"EVdata/proto_struct"
	"encoding/csv"
	"os"
)

var PointHeaders = []string{
	"id",
	"time",
	"speed",
	"longitude",
	"latitude",
	"vehicle_status",
	"have_driver",
	"have_brake",
	"accelerator_pedal",
	"brake_status",
	"matched_lon",
	"matched_lat",
	"road_id",
	"track_id",
}

type PointWriter struct {
	csvWriter *csv.Writer
	file      *os.File
}

func NewPointWriter(path string) *PointWriter {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0744)
	if err != nil {
		common.ErrorLog("Failed to create file:", err)
	}

	w := &PointWriter{file: f, csvWriter: csv.NewWriter(f)}
	if err := w.csvWriter.Write(PointHeaders); err != nil {
		common.ErrorLog("Error when writing header", err)
	}
	return w
}

func (w *PointWriter) Write(data []*proto_struct.TrackPoint) {
	if data == nil || len(data) == 0 {
		common.ErrorLog("data is nil")
		return
	}

	for _, p := range data {
		if err := w.csvWriter.Write(p.ToCSV()); err != nil {
			common.ErrorLog("Error when writing data", err, p)
		}
	}
	w.csvWriter.Flush()
	_ = w.file.Close()
}
