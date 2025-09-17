package proto_struct

import "fmt"

func (p *RawPoint) ToCSV() []string {
	return []string{
		//fmt.Sprint(p.Vin),
		p.Time,
		fmt.Sprint(p.Speed),
		fmt.Sprint(p.Longitude),
		fmt.Sprint(p.Latitude),
		fmt.Sprint(p.VehicleStatus),
		fmt.Sprint(p.HaveDriver),
		fmt.Sprint(p.HaveBrake),
		fmt.Sprint(p.AcceleratorPedal),
		fmt.Sprint(p.BrakeStatus),
	}
}

var MatchingPointHeader = []string{
	//"longitude",
	//"latitude",
	"matched_lon",
	"matched_lat",
	"road_id",
	"node_id",
	"is_bad",
}

func (p *MatchingPoint) ToCSV() []string {
	return []string{
		//fmt.Sprintf("%.6f", p.OriginalLon),
		//fmt.Sprintf("%.6f", p.OriginalLat),
		fmt.Sprintf("%.7f", p.MatchedLon),
		fmt.Sprintf("%.7f", p.MatchedLat),
		fmt.Sprint(p.RoadId),
		fmt.Sprint(p.NodeId),
		fmt.Sprint(p.IsBad),
	}
}

var TrackPointHeader = []string{
	"time",
	"speed",
	"longitude",
	"latitude",
	"vehicle_status",
	"have_driver",
	"have_brake",
	"accelerator_pedal",
	"brake_status",
	"matched_lon",
	"matched_lat",
	"road_id",
	"node_id",
	"is_bad",
}

func (p *TrackPoint) ToCSV() []string {
	return []string{
		p.Time,
		fmt.Sprint(p.Speed),
		fmt.Sprintf("%.6f", p.Longitude),
		fmt.Sprintf("%.6f", p.Latitude),
		fmt.Sprint(p.VehicleStatus),
		fmt.Sprint(p.HaveDriver),
		fmt.Sprint(p.HaveBrake),
		fmt.Sprint(p.AcceleratorPedal),
		fmt.Sprint(p.BrakeStatus),

		fmt.Sprintf("%.6f", p.MatchedLon),
		fmt.Sprintf("%.6f", p.MatchedLat),
		fmt.Sprint(p.RoadId),
		fmt.Sprint(p.IsBad),
	}
}
