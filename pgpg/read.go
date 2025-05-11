package pgpg

import (
	"EVdata/common"
	"EVdata/proto_struct"
	"fmt"
	"github.com/parquet-go/parquet-go"
	"log"
	"os"
)

func GetMetaData1(filename string) {
	cfg, err := parquet.NewReaderConfig()
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := parquet.NewReader(file, cfg)

	metaFile, err := os.OpenFile("metadata.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0744)
	if err != nil {
		log.Fatal(err)
	}
	defer metaFile.Close()

	_, err = metaFile.WriteString(fmt.Sprint(reader.Schema()))
	if err != nil {
		log.Fatal(err)
	}
}

func ReadPointFromParquet(filename string) []*common.RawPoint {
	cfg, err := parquet.NewReaderConfig()
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := parquet.NewReader(file, cfg)

	numRows := reader.NumRows()
	rows := make([]*common.RawPoint, numRows)
	var i int64
	for i = 0; i < numRows; i++ {
		var vehicle common.RawPoint
		err = reader.Read(&vehicle)
		if err != nil {
			log.Println("error when reading data", err)
		}
		rows[i] = &vehicle
	}
	return rows
}

func ReadTrackPointFromParquet(filename string) []*proto_struct.TrackPoint {
	//start := time.Now()
	cfg, err := parquet.NewReaderConfig()
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.Open(filename)
	if err != nil {
		common.InfoLog(filename, err)
		return nil
	}
	defer file.Close()

	reader := parquet.NewReader(file, cfg)

	numRows := reader.NumRows()
	//log.Println(reader.Schema(), numRows)
	rows := make([]*proto_struct.TrackPoint, numRows)
	var i int64
	for i = 0; i < numRows; i++ {
		v := common.GetTrackPoint(nil)
		err = reader.Read(v)
		if err != nil {
			common.ErrorLog("error when reading data", err)
		}
		rows[i] = v
	}
	//common.InfoLog("time for reading file", time.Since(start))
	return rows
}

func ReadTrackFromParquet(filename string) []*common.Track {
	common.InfoLog("reading file", filename)
	//start := time.Now()
	cfg, err := parquet.NewReaderConfig()
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := parquet.NewReader(file, cfg)

	numRows := reader.NumRows()
	rows := make([]*common.Track, numRows)
	var i int64
	for i = 0; i < numRows; i++ {
		var vehicle common.Track
		err = reader.Read(&vehicle)
		if err != nil {
			common.ErrorLog("error when reading data", err)
		}
		rows[i] = &vehicle
	}
	//common.InfoLog("time for reading file", time.Since(start))
	return rows
}

func ReadFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	st, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("start reading file")
	rows, err := parquet.Read[common.TrackPoint](file, st.Size())
	log.Println(rows[0], len(rows))
	log.Println("finished reading file")
	if err != nil {
		log.Println(err)
	}

	//log.Println(rows[0])

	//for _, row := range rows {
	//	log.Println(row)
	//}
}
