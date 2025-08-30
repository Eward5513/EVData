package CSV

import (
	"EVdata/common"
	"EVdata/proto_struct"
	"encoding/csv"
	"os"
)

var pointHeaders = []string{
	"longitude",
	"latitude",
	"matched_lon",
	"matched_lat",
	"road_id",
	"is_bad",
}

type MatchingPointWriter struct {
	csvWriter *csv.Writer
	file      *os.File
}

func NewMatchingPointWriter(path string) *MatchingPointWriter {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0744)
	if err != nil {
		common.ErrorLog("Failed to create file:", err)
	}

	w := &MatchingPointWriter{file: f, csvWriter: csv.NewWriter(f)}
	if err := w.csvWriter.Write(pointHeaders); err != nil {
		common.ErrorLog("Error when writing header", err)
	}
	return w
}

func (w *MatchingPointWriter) Write(data []*proto_struct.MatchingPoint) {
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
