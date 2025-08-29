package proto_struct

import "fmt"

func (p *TrackPoint) ToCSV() []string {
	return []string{
		fmt.Sprint(p.Id),
		p.Time,
		fmt.Sprint(p.Speed),
		fmt.Sprint(p.Longitude),
		fmt.Sprint(p.Latitude),
		fmt.Sprint(p.VehicleStatus),
		fmt.Sprint(p.HaveDriver),
		fmt.Sprint(p.HaveBrake),
		fmt.Sprint(p.AcceleratorPedal),
		fmt.Sprint(p.BrakeStatus),
		fmt.Sprint(p.MatchedLon),
		fmt.Sprint(p.MatchedLat),
		fmt.Sprint(p.RoadId),
		fmt.Sprint(p.TrackId),
	}
}
