package pgpg

import (
	"EVdata/common"
	"bufio"
	"github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/compress/zstd"
	"log"
	"os"
)

var (
	WriterBuffer = 10
)

type PQTrackWriter struct {
	fileHandle *os.File
	filename   string
	gw         *parquet.GenericWriter[*common.Track]
	data       []*common.Track
}

func NewPQTrackWriter(f string) *PQTrackWriter {
	comp := parquet.Compression(&zstd.Codec{})
	bfsize := parquet.WriteBufferSize(0)
	cfg, err := parquet.NewWriterConfig(comp, bfsize)
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Create(f)
	if err != nil {
		log.Fatal(err)
	}
	wb := bufio.NewWriter(file)

	ww := parquet.NewGenericWriter[*common.Track](wb, cfg)
	return &PQTrackWriter{filename: f, fileHandle: file, gw: ww, data: make([]*common.Track, 0, WriterBuffer)}
}

func (w *PQTrackWriter) Write(p *common.Track) {
	w.data = append(w.data, p)
	if len(w.data) >= WriterBuffer {
		if _, err := w.gw.Write(w.data); err != nil {
			common.ErrorLog("PQTrackWriter.Write()", err)
		}
		if err := w.gw.Flush(); err != nil {
			common.ErrorLog("PQTrackWriter.Flush()", err)
		}
		//for _, t := range w.data {
		//	common.PutTrack(t)
		//}
		w.data = w.data[:0]
		//common.InfoLog("flush data")
	}
}

func (w *PQTrackWriter) Close() {
	defer w.fileHandle.Close()

	if _, err := w.gw.Write(w.data); err != nil {
		common.ErrorLog("error when flush parquet file", err)
	}
	if err := w.gw.Flush(); err != nil {
		common.ErrorLog("error when flush parquet file", err)
	}
	if err := w.gw.Close(); err != nil {
		common.ErrorLog("error when close parquet file", err)
	}
	//common.InfoLog("Finish writing file", w.filename)
}
