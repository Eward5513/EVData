package common

import (
	"encoding/json"
	"fmt"
	"log"
)

type RawPoint struct {
	Vin            string  `parquet:"vin"`
	CollectionTime int64   `parquet:"collectiontime"`
	Speed          float64 `parquet:"speed"`
	LocationState  int32   `parquet:"locationstate"`
	Longitude      float64 `parquet:"longitude"`
	Latitude       float64 `parquet:"latitude"`
	//ReceivedTime    int64               `parquet:"receivetime"`
	//VehicleData     VehicleDataType     `parquet:"vehicledata"`
	//VehiclePosition VehiclePositionType `parquet:"vehicleposition"`
	//EngineData      EngineDataType      `parquet:"enginedata"`
	//ExtremeValue    ExtremeValueType    `parquet:"extremevalue"`
	//FuelCell        FuelCellType        `parquet:"fuelcell"`
	//MotorDataList   string              `parquet:"motordatalist"`
	TimeStamp int
}

func (data *RawPoint) String() string {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Println("Error marshalling vehicle ", err)
		return ""
	}
	return string(jsonData)
}

func (data *RawPoint) IsValid() bool {
	//return true
	if data.Speed > 220 {
		data.Speed = -1
	}
	return data.LocationState == 0 &&
		data.Latitude >= MIN_LATITUDE &&
		data.Latitude <= MAX_LATITUDE &&
		data.Longitude >= MIN_LONGITUDE &&
		data.Longitude <= MAX_LONGITUDE
}

type VehicleDataType struct {
	VehicleStatus        int32   `parquet:"vehiclestatus"`
	ChargeStatus         int32   `parquet:"chargestatus"`
	RunModel             int32   `parquet:"runmodel"`
	Speed                float64 `parquet:"speed"`
	SumMileage           float64 `parquet:"summileage"`
	SumVoltage           float64 `parquet:"sumvoltage"`
	SumCurrent           float64 `parquet:"sumcurrent"`
	Soc                  int32   `parquet:"soc"`
	DcdcStatus           int32   `parquet:"dcdcstatus"`
	GearNum              int32   `parquet:"gearnum"`
	HaveDriver           int32   `parquet:"havedriver"`
	HaveBrake            int32   `parquet:"havebrake"`
	InsulationResistance int32   `parquet:"insulationresistance"`
	AcceleratorPedal     int32   `parquet:"acceleratorpedal"`
	BrakeStatus          int32   `parquet:"brakestatus"`
}

func (data *VehicleDataType) ToCsv() []string {
	return []string{
		fmt.Sprintf("%d", data.VehicleStatus),
		fmt.Sprintf("%d", data.ChargeStatus),
		fmt.Sprintf("%d", data.RunModel),
		fmt.Sprintf("%f", data.Speed),
		fmt.Sprintf("%f", data.SumMileage),
		fmt.Sprintf("%f", data.SumVoltage),
		fmt.Sprintf("%f", data.SumCurrent),
		fmt.Sprintf("%d", data.Soc),
		fmt.Sprintf("%d", data.DcdcStatus),
		fmt.Sprintf("%d", data.GearNum),
		fmt.Sprintf("%d", data.HaveDriver),
		fmt.Sprintf("%d", data.HaveBrake),
		fmt.Sprintf("%d", data.InsulationResistance),
		fmt.Sprintf("%d", data.AcceleratorPedal),
		fmt.Sprintf("%d", data.BrakeStatus),
	}
}

type VehiclePositionType struct {
	LocationState int32   `parquet:"locationstate"`
	Lat           int32   `parquet:"Lat"`
	Lon           int32   `parquet:"Lon"`
	Longitude     float64 `parquet:"longitude"`
	Latitude      float64 `parquet:"latitude"`
}

func (data *VehiclePositionType) ToCsv() []string {
	return []string{
		//fmt.Sprintf("%d", data.LocationState),
		//fmt.Sprintf("%d", data.Lat),
		//fmt.Sprintf("%d", data.Lon),
		fmt.Sprintf("%.7f", data.Longitude),
		fmt.Sprintf("%.7f", data.Latitude),
	}
}

type EngineDataType struct {
	EngineStatus              int32   `parquet:"enginestatus"`
	GrankshaftSpeed           int32   `parquet:"grankshaftspeed"`
	EngineFuelConsumptionrate float64 `parquet:"enginefuelconsumptionrate"`
}

func (data *EngineDataType) ToCsv() []string {
	return []string{
		fmt.Sprintf("%d", data.EngineStatus),
		fmt.Sprintf("%d", data.GrankshaftSpeed),
		fmt.Sprintf("%f", data.EngineFuelConsumptionrate),
	}
}

type ExtremeValueType struct {
	MaxVoltageBatterySubSeq           int32   `parquet:"maxvoltagebatterysubseq"`
	MaxVoltageBatterySingleSeq        int32   `parquet:"maxvoltagebatterysingleseq"`
	MaxBatterySingleVoltageVal        float64 `parquet:"maxbatterysinglevoltageval"`
	MinVoltageBatterySubSeq           int32   `parquet:"minvoltagebatterysubseq"`
	MinVoltageBatterySingleVoltageVal float64 `parquet:"minbatterysinglevoltageval"`
	MinBatterySingleVoltageVal        float64 `parquet:"minbatterysinglevoltageval"`
	MaxTmpSubSeq                      int32   `parquet:"maxtmpsubseq"`
	MaxTmpprobeSingleSeq              int32   `parquet:"maxtmpprobesingleseq"`
	MaxTmpVal                         int32   `parquet:"maxtmpval"`
	MinTmpSubSeq                      int32   `parquet:"mintmpsubseq"`
	MinTmpprobeSingleSeq              int32   `parquet:"mintmpprobesingleseq"`
	MinTmpVal                         int32   `parquet:"mintmpval"`
}

func (data *ExtremeValueType) ToCsv() []string {
	return []string{
		fmt.Sprintf("%d", data.MaxVoltageBatterySubSeq),
		fmt.Sprintf("%d", data.MaxVoltageBatterySingleSeq),
		fmt.Sprintf("%f", data.MaxBatterySingleVoltageVal),
		fmt.Sprintf("%d", data.MinVoltageBatterySubSeq),
		fmt.Sprintf("%f", data.MinVoltageBatterySingleVoltageVal),
		fmt.Sprintf("%f", data.MinBatterySingleVoltageVal),
		fmt.Sprintf("%d", data.MaxTmpSubSeq),
		fmt.Sprintf("%d", data.MaxTmpprobeSingleSeq),
		fmt.Sprintf("%d", data.MaxTmpVal),
		fmt.Sprintf("%d", data.MinTmpSubSeq),
		fmt.Sprintf("%d", data.MinTmpprobeSingleSeq),
		fmt.Sprintf("%d", data.MinTmpVal),
	}
}

type FuelCellType struct {
	FuelCellVoltage             float64 `parquet:"fuelcellvoltage"`
	FuelCellElectric            float64 `parquet:"fuelcellelectric"`
	FuelConsumption             float64 `parquet:"fuelconsumption"`
	FuelCellTmpprobeNumber      int32   `parquet:"fuelcelltmpprobenumber"`
	ProbeTmpValues              string  `parquet:"probetmpvalues"`
	H2MaxTmp                    float64 `parquet:"h2maxtmp"`
	H2MaxTmpprobeSeq            int32   `parquet:"h2maxtmpprobeseq"`
	H2MaxConcentration          int32   `parquet:"h2maxconcentration"`
	H2MaxConcentrationSensorSeq int32   `parquet:"h2maxconcentrationsensorseq"`
	H2MaxPressure               float64 `parquet:"h2maxpressure"`
	H2MaxPressureSensorSeq      int32   `parquet:"h2maxpressuresensorseq"`
	HighVoltageDcdState         int32   `parquet:"highvoltagedcdstate"`
}

func (data *FuelCellType) ToCsv() []string {
	return []string{
		fmt.Sprintf("%f", data.FuelCellVoltage),
		fmt.Sprintf("%f", data.FuelCellElectric),
		fmt.Sprintf("%f", data.FuelConsumption),
		fmt.Sprintf("%d", data.FuelCellTmpprobeNumber),
		data.ProbeTmpValues,
		fmt.Sprintf("%f", data.H2MaxTmp),
		fmt.Sprintf("%d", data.H2MaxTmpprobeSeq),
		fmt.Sprintf("%d", data.H2MaxConcentration),
		fmt.Sprintf("%d", data.H2MaxConcentrationSensorSeq),
		fmt.Sprintf("%f", data.H2MaxPressure),
		fmt.Sprintf("%d", data.H2MaxPressureSensorSeq),
		fmt.Sprintf("%d", data.HighVoltageDcdState),
	}
}

// 用于排序
type VehicleArray []*RawPoint

func (va VehicleArray) Len() int           { return len(va) }
func (va VehicleArray) Swap(i, j int)      { va[i], va[j] = va[j], va[i] }
func (va VehicleArray) Less(i, j int) bool { return va[i].CollectionTime < va[j].CollectionTime }

var VehicleColumns = []string{
	"vin",
	"collection_time",
	"received_time",
	"vehicle_status",
	"charge_status",
	"run_model",
	"speed",
	"sum_mileage",
	"sum_voltage",
	"sum_current",
	"soc",
	"dcdc_status",
	"gear_num",
	"have_driver",
	"have_brake",
	"insulation_resistance",
	"accelerator_pedal",
	"brake_status",
	"location_state",
	"Lat",
	"Lon",
	"longitude",
	"latitude",
	"engine_status",
	"grankshaft_speed",
	"engine_fuel_consumptionrate",
	"max_voltage_battery_sub_seq",
	"max_voltage_battery_single_seq",
	"max_battery_single_voltage_val",
	"min_voltage_battery_sub_seq",
	"min_battery_single_voltage_val",
	"min_battery_single_voltage_val",
	"max_tmp_sub_seq",
	"max_tmp_probe_single_seq",
	"max_tmp_val",
	"min_tmp_sub_seq",
	"min_tmp_probe_single_seq",
	"min_tmp_val",
	"fuel_cell_voltage",
	"fuel_cell_electric",
	"fuel_consumption",
	"fuel_cell_tmp_probe_number",
	"probe_tmp_values",
	"h2_max_tmp",
	"h2_max_tmp_probe_seq",
	"h2_max_concentration",
	"h2_max_concentration_sensor_seq",
	"h2_max_pressure",
	"h2_max_pressure_sensor_seq",
	"high_voltage_dcd_state",
}

var PivotalVehicleColumns = []string{
	"vin",
	"collection_time",
	"date",
	"timestamp",
	"hour",
	"speed",
	"longitude",
	"latitude",
}
