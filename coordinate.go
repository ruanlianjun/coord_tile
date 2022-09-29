package coordtile

import "math"

type (
	Coordinate struct {
		X     float64
		Y     float64
		Level int
	}
)

func (c Coordinate) WGS84ToWebMercator() Coordinate {
	var result Coordinate
	result.X = c.X * 20037508.34 / 180
	result.Y = math.Log(math.Tan((90+c.Y)*math.Pi/360)) / (math.Pi / 180)
	result.Y = result.Y * 20037508.34 / 180
	return result
}

func (c Coordinate) WebMercatorToWGS84() Coordinate {
	var result Coordinate
	result.X = c.X / 20037508.34 * 180
	result.Y = c.Y / 20037508.34 * 180
	result.Y = 180 / math.Pi * (2*math.Atan(math.Exp(result.Y*math.Pi/180)) - math.Pi/2)
	return result
}
