package pgpg

import (
	"EVdata/common"
	"EVdata/proto_struct"
	"github.com/parquet-go/parquet-go"
	"log"
	"os"
)

type PQPointWriter struct {
	data     []*proto_struct.TrackPoint
	filename string
}

func NewPQWriter(f string) *PQPointWriter {
	return &PQPointWriter{filename: f, data: make([]*proto_struct.TrackPoint, 0)}
}

func (w *PQPointWriter) Write(ps []*proto_struct.TrackPoint) {
	w.data = append(w.data, ps...)
}

func (w *PQPointWriter) Close() {
	if len(w.data) == 0 {
		common.InfoLog("PQWriter: no data")
		return
	}

	cfg, err := parquet.NewWriterConfig()
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.Create(w.filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	ww := parquet.NewGenericWriter[*proto_struct.TrackPoint](file, cfg)
	if _, err = ww.Write(w.data); err != nil {
		common.ErrorLog("error when write parquet file", err)
	}
	if err = ww.Flush(); err != nil {
		common.ErrorLog("error when flush parquet file", err)
	}
	if err = ww.Close(); err != nil {
		common.ErrorLog("error when close parquet file", err)
	}
}
