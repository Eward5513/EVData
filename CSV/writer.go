package CSV

import (
	"EVdata/common"
	"encoding/csv"
	"os"
)

func Write(fp string, header []string, data [][]string) {
	f, err := os.OpenFile(fp, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0744)
	if err != nil {
		common.ErrorLog("Failed to create file:", err)
	}

	if data == nil || len(data) == 0 {
		common.ErrorLog("data is nil")
		return
	}

	csvWriter := csv.NewWriter(f)
	if err = csvWriter.Write(header); err != nil {
		common.ErrorLog("Error when writing header", err)
	}

	for _, p := range data {
		if err = csvWriter.Write(p); err != nil {
			common.ErrorLog("Error when writing data", err, p)
		}
	}
	csvWriter.Flush()
	_ = f.Close()
}
