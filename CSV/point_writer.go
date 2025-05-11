package CSV

import (
	"EVdata/common"
	"EVdata/proto_struct"
	"encoding/csv"
	"errors"
	"log"
	"os"
	"sort"
)

type PointWriter struct {
	csvWriter *csv.Writer
	file      *os.File
	data      []*proto_struct.TrackPoint
}

func NewPointWriter(f *os.File) *PointWriter {
	w := &PointWriter{file: f, csvWriter: csv.NewWriter(f)}
	w.data = make([]*proto_struct.TrackPoint, 0)
	if err := w.csvWriter.Write(common.PivotalVehicleColumns); err != nil {
		log.Println("Error when writing header", err)
	}
	return w
}

func (w *PointWriter) Write(d *proto_struct.TrackPoint) error {
	if d == nil {
		log.Println("data is nil")
		return errors.New("data is nil")
	}
	w.data = append(w.data, d)
	return nil
}

func (w *PointWriter) Close() {
	sort.Slice(w.data, func(i, j int) bool { return (w.data[i]).CollectionTime < (w.data[j]).CollectionTime })
	for _, d := range w.data {
		if err := w.csvWriter.Write(d.ToCsv()); err != nil {
			log.Println("Error when writing data", err)
		}
	}
	w.csvWriter.Flush()
	_ = w.file.Close()
}

func WriteSummary(data [][]string, fileName string) {
	f, err := os.Create(fileName)
	if err != nil {
		log.Println("Error when writing summary", err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	if err := w.WriteAll(data); err != nil {
		log.Println("Error when writing summary", err)
	}
}
