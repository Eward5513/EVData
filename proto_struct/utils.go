package proto_struct

import "fmt"

func (p *RawPoint) ToCSV() []string {
	return []string{
		fmt.Sprint(p.Vin),
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

func (p *MatchingPoint) ToCSV() []string {
	return []string{
		fmt.Sprintf("%.6f", p.OriginalLon),
		fmt.Sprintf("%.6f", p.OriginalLat),
		fmt.Sprintf("%.6f", p.MatchedLon),
		fmt.Sprintf("%.6f", p.MatchedLat),
		fmt.Sprint(p.RoadId),
		fmt.Sprint(p.IsBad),
	}
}
