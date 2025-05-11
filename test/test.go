package main

import (
	"EVdata/common"
	"github.com/parquet-go/parquet-go"
	"os"
	"time"
)

func main() {
	time.Local = time.UTC
	w := parquet.NewGenericWriter[*common.Track](os.Stdout)
	w.Write([]*common.Track{})
}
