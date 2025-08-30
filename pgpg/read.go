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

func ReadPointFromParquet(filename string) [][]*proto_struct.RawPoint {
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
	rows := make([][]*proto_struct.RawPoint, common.VEHICLE_COUNT+1)
	var i int64
	for i = 0; i < numRows; i++ {
		vehicle := common.GetRawPoint()
		err = reader.Read(vehicle)
		vehicle.TimeInt = common.ParseTimeToInt(vehicle.Time)
		if err != nil {
			log.Println("error when reading data", err)
		}
		rows[vehicle.Vin] = append(rows[vehicle.Vin], vehicle)
	}
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
	rows, err := parquet.Read[common.RawPoint](file, st.Size())
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
