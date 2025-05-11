package xp

import (
	"encoding/csv"
	"fmt"
	"github.com/xitongsys/parquet-go-source/buffer"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/reader"
	"log"
	"os"
)

type Type struct {
	//BigIntCol int64 `parquet:"NAME=bigint_col,type=BIG_INT"`
	//BoolCol       bool             `parquet:"bool_col"`
	//DateStringCol string           `parquet:"date_string_col"`
	//DoubleCol     float64          `parquet:"double_col"`
	//FloatCol      float32          `parquet:"float_col"`
	//Id            int32            `parquet:"id"`
	//IntCol        int32            `parquet:"int_col"`
	//SmallintCol   int32            `parquet:"smallint_col"`
	StringCol string `parquet:"name=vin, type=BYTE_ARRAY"`
	//TimestampCol  deprecated.Int96 `parquet:"timestamp_col"`
	//TinyintCol    int32            `parquet:"tinyint_col"`
}

func GetMetaData(filename string) {
	fr, err := local.NewLocalFileReader(filename)
	if err != nil {
		log.Println("Can't open file", err.Error())
		return
	}
	defer fr.Close()
	buf := make([]byte, 1024*1024*1024*4)
	file, err := os.Open(filename)
	log.Println("1111111111")
	_, err = file.Read(buf)
	if err != nil {
		log.Println("Can't read file", err.Error())
	}
	log.Println("22222222222222")
	bfile, err := buffer.NewBufferFile(buf)
	if err != nil {
		log.Println("Can't create buffer file", err.Error())
	}

	log.Println("33333333333333")
	log.Println(bfile)
	pr, err := reader.NewParquetReader(fr, new(Type), 1)

	if err != nil {
		log.Println("Can't create parquet reader", err)
		return
	}

	var data Type
	log.Println("1111111")
	if err := pr.Read(&data); err != nil {
		log.Println("Can't read parquet file", err)
	}
	log.Println(data)

	var output *os.File
	os.Remove("output.CSV")
	if output, err = os.Create("output.CSV"); err != nil {
		fmt.Println("Can't create file", err)
	}
	defer output.Close()

	csvWriter := csv.NewWriter(output)
	defer csvWriter.Flush()
	if err := csvWriter.Write([]string{"Name", "Type", "LogicalType", "RepetitionType", "NumChildren"}); err != nil {
		log.Println("Can't write to file", err)
	}
	for _, sc := range pr.Footer.Schema {
		if err := csvWriter.Write([]string{fmt.Sprint(sc.Name), fmt.Sprint(sc.GetType()), fmt.Sprint(getLogicTye(sc.LogicalType)), fmt.Sprint(sc.RepetitionType), fmt.Sprint(sc.GetNumChildren())}); err != nil {
			log.Fatal(err)
		}
	}
}

func getLogicTye(p *parquet.LogicalType) string {
	var ans string
	if p == nil {
		return ans
	}
	if p.IsSetSTRING() {
		ans += "STRING "
	}
	if p.IsSetMAP() {
		ans += "MAP "
	}
	if p.IsSetLIST() {
		ans += "LIST "
	}
	if p.IsSetENUM() {
		ans += "ENUM "
	}
	if p.IsSetDECIMAL() {
		ans += "DECIMAL "
	}
	if p.IsSetDATE() {
		ans += "DATE "
	}
	if p.IsSetTIME() {
		ans += "TIME "
	}
	if p.IsSetTIMESTAMP() {
		ans += "TIMESTAMP "
	}
	if p.IsSetINTEGER() {
		ans += "INTEGER "
	}
	if p.IsSetUNKNOWN() {
		ans += "UNKNOWN "
	}
	if p.IsSetJSON() {
		ans += "JSON "
	}
	if p.IsSetBSON() {
		ans += "BSON "
	}
	if p.IsSetUUID() {
		ans += "UUID "
	}
	return ans
}
