package main

import (
	"EVdata/common"
	"EVdata/pgpg"
	"log"
	"os"
	"path/filepath"
)

func main() {
	common.CreatLogFile("summary.log")
	defer common.CloseLogFile()

	SummarizePoint()

}

func SummarizePoint() {
	basePath := common.REFINED_RAW_DATA_DIR_PATH
	fs, err := os.ReadDir(basePath)
	if err != nil {
		log.Fatal("Unable to read top dir", err.Error())
	}

	var cnt int
	for i, f := range fs {
		if filepath.Ext(f.Name()) == ".parquet" {
			rows := pgpg.ReadTrackPointFromParquet(filepath.Join(basePath, f.Name()))
			cnt += len(rows)
		}
		common.InfoLog("Finish reading ", i)
	}
	common.InfoLog("Total points: ", cnt)
}
