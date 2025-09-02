package traffic_flow

import (
	"EVdata/CSV"
	"EVdata/common"
	"EVdata/proto_struct"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	workerCount int = 500
)

type mergeData struct {
	rps []*proto_struct.RawPoint
	mps []*proto_struct.MatchingPoint
}

func StartReader(rch chan *mergeData) {
	wg := &sync.WaitGroup{}
	rrch := make(chan int, workerCount)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for vin := range rrch {
				rps := CSV.ReadRawPointFromCSV(filepath.Join(common.RAW_DATA_DIR_PATH, fmt.Sprint(vin)+".csv"), vin)
				mps := make([]*proto_struct.MatchingPoint, 0, len(rps))
				for j := 0; j < 100; j++ {
					fp := filepath.Join(common.RAW_DATA_DIR_PATH, fmt.Sprintf("%d_%d.csv", vin, j))
					_, err := os.Stat(fp)
					if err != nil {
						break
					}
					ps := CSV.ReadMatchingPointFromCSV(fp)
					mps = append(mps, ps...)
				}

				rch <- &mergeData{rps: rps, mps: mps}

			}
		}()
	}

	for i := 1; i < common.VEHICLE_COUNT; i++ {
		rrch <- i
	}
	close(rrch)
	wg.Wait()

}

func StartWriter(wch chan []*proto_struct.TrackPoint) {
	wg := &sync.WaitGroup{}
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for tps := range wch {
				if tps == nil || len(tps) == 0 {
					common.ErrorLog("empty trackpoints")
					continue
				}
				vin := tps[0].Vin
				fp := filepath.Join(common.TRCK_POINT_CSV_DIR, fmt.Sprint(vin)+".csv")
				//CSV.WriteTrackPoints(fp, tps)
			}
		}()
	}
	wg.Wait()
}

func StartWorker(rch chan *mergeData, wch chan []*proto_struct.TrackPoint) {
	wg := &sync.WaitGroup{}
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for data := range rch {
				res := make([]*proto_struct.TrackPoint, 0, len(data.rps)+len(data.mps))
				for j := 0; j < len(data.rps); {
					for k := 0; k < len(data.mps); {
						if data.rps[j].Longitude == data.mps[k].OriginalLon && data.rps[j].Latitude == data.mps[k].OriginalLat {
							p := &proto_struct.TrackPoint{
								Vin:              data.rps[j].Vin,
								Speed:            data.rps[j].Speed,
								Longitude:        data.rps[j].Longitude,
								Latitude:         data.rps[j].Latitude,
								HaveBrake:        data.rps[j].HaveBrake,
								HaveDriver:       data.rps[j].HaveDriver,
								AcceleratorPedal: data.rps[j].AcceleratorPedal,
							}
						}
					}
				}
			}
		}()
	}

	wg.Wait()
	close(wch)
}
