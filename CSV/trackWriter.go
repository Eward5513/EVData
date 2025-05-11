package CSV

import (
	"EVdata/common"
	"EVdata/proto_struct"
	"encoding/csv"
	"os"
)

type TrackWriter struct {
	csvWriter *csv.Writer
	file      *os.File
	data      *proto_struct.Track
}

func NewTrackWriter(n string) *TrackWriter {
	f, err := os.Create(n)
	if err != nil {
		common.ErrorLog(err)
	}
	w := &TrackWriter{file: f, csvWriter: csv.NewWriter(f)}
	if err := w.csvWriter.Write(common.TRACK_HEADER); err != nil {
		common.ErrorLog("Error when writing header", err)
	}
	return w
}

func (w *TrackWriter) Write(t *proto_struct.Track) {
	if t == nil {
		common.ErrorLog("data is nil")
	}
	w.data = t
}

func (w *TrackWriter) Close() {
	if err := w.csvWriter.WriteAll(w.data.ToCsv()); err != nil {
		common.ErrorLog("Error when writing data", err)
	}
	w.csvWriter.Flush()
	//common.InfoLog("Data written to ", w.file.Name())
	_ = w.file.Close()
}
