package main

import (
	"EVdata/CSV"
	"EVdata/common"
	"EVdata/pgpg"
	"EVdata/proto_tools"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func main() {

	begin := time.Now()

	common.CreatLogFile("validate.log")
	defer common.CloseLogFile()

	var pointID, trackID int
	flag.IntVar(&pointID, "p", -1, "point mode")
	flag.IntVar(&trackID, "t", -1, "track mode")
	flag.Parse()

	if pointID > 0 {
		validatePoint(pointID)
	}
	if trackID > 0 {
		validateTrack(trackID)
	}

	common.InfoLog("Execution time: ", time.Now().Sub(begin))
}

func validatePoint(pointID int) {
	basePath := common.REFINED_RAW_DATA_DIR_PATH

	wf, err := os.Create("validate_point.csv")
	if err != nil {
		log.Fatal("Unable to create file", err.Error())
	}
	w := CSV.NewMatchingPointWriter(wf)
	rows := pgpg.ReadTrackPointFromParquet(filepath.Join(basePath, strconv.Itoa(pointID)+".prob"))
	log.Println(len(rows), rows[0].Vin)
	for _, row := range rows {
		_ = w.Write(row)
	}
	w.Close()
	common.InfoLog("Finish reading ")
	return
}

func validateTrack(trackID int) {
	basePath := common.TRACK_RAW_DATA_DIR_PATH
	dirName := "validate_track"
	if err := os.RemoveAll(dirName); err != nil {
		common.FatalLog("Unable to remove dir: ", dirName)
	}
	if err := os.Mkdir(dirName, 0755); err != nil {
		common.FatalLog("Unable to create dir: ", dirName)
	}

	ts := proto_tools.ReadTrackFromProto(filepath.Join(basePath, strconv.Itoa(trackID)+".prob"))
	for _, t := range ts {
		w := CSV.NewTrackWriter(filepath.Join(dirName, fmt.Sprintf("%s_%d_%d.csv", t.Date, t.Vin, t.Tid)))
		log.Println(t.Vin, t.Date, t.Tid, t.StartTime, t.EndTime, len(t.TrackSegs))
		w.Write(t)
		w.Close()
	}

	common.InfoLog("Finish reading ")
	//pgpg.ReadFile(filepath.Join(basePath, "1.parquet"))
	return
}
