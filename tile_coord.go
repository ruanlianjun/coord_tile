package coordtile

import (
	"math"

	"golang.org/x/sync/errgroup"
)

// TileProjection 投影
type TileProjection int

const (
	TileProjWGS1984     TileProjection = iota // EPSG:4326，[-180.0, 180.0], [-85.0511288, 85.0511288]
	TileProjCGCS2000                          // EPSG:4490, [-180.0, 180.0], [-90, 90]
	TileProjWebMercator                       // EPSG:3857, [-20037508.3427892, 20037508.3427892], [-20037508.3427892, 20037508.3427892]
	TileProjTianDiTu                          // EPSG:4490, [-180.0, 180.0], [-90, 90]
	TileProjArcGIS                            // EPSG:3857, [-20037508.342787, 20037508.342787], [-19971868.88040859, 19971868.88040859]
	TileProjBaidu                             // 参考椭球：Clarke_1866，坐标系：NAD27， 投影：Mercator
)

type TileResolutionUnit int

const (
	Meter TileResolutionUnit = iota
	Degree
)

type TileCoord struct {
	Projection        TileProjection
	TileSize          int     // 瓦片大小, 目前只支持256×256大小的瓦片
	OriX              float64 // 水平方向原点，单位：度或者米
	OriY              float64
	StartLevel        int     // 起始等级，例如0
	initialResolution float64 //  起始等级对应的分辨率, 两个方向的分辨率相同，单位：度每像素或者米每像素。
	TileUnit          TileResolutionUnit
}

type (
	TileCoordinate struct {
		X, Y, Level int
	}
	TileCoordinateBound []TileCoordinate
	TileCoordinateScope struct {
		MinX, MaX, MinY, MaY, Level int
	}
)

func NewTileCoord(projection TileProjection) *TileCoord {
	ts := &TileCoord{
		Projection: projection,
		TileSize:   256,
	}

	switch projection {
	case TileProjWGS1984, TileProjCGCS2000, TileProjTianDiTu:
		ts.OriX = -180.0
		ts.OriY = 90.0
		ts.StartLevel = 1
		ts.initialResolution = 0.703125
		ts.TileUnit = Degree
	case TileProjWebMercator, TileProjArcGIS:
		ts.OriX = -20037508.342787
		ts.OriY = 20037508.342787
		ts.StartLevel = 0
		ts.initialResolution = 156543.03392798
		ts.TileUnit = Meter
	case TileProjBaidu:
		ts.OriX = 0
		ts.OriY = 0
		ts.StartLevel = 0
		ts.initialResolution = 0
		ts.TileUnit = Meter
	default:
		return nil
	}
	return ts
}

func (t *TileCoord) Resolution(level int) float64 {
	max := 1 << (level - t.StartLevel)
	return t.initialResolution / float64(max)
}

func (t *TileCoord) CalcTileCoordinate(c Coordinate, level int) TileCoordinate {
	var resp TileCoordinate
	resp.X = int((c.X - t.OriX) / t.Resolution(level) / float64(t.TileSize))
	resp.Y = int(math.Abs(t.OriY-c.Y) / t.Resolution(level) / float64(t.TileSize))
	resp.Level = level
	return resp
}

func (t *TileCoord) WebMercatorTile(c Coordinate, level int) TileCoordinate {
	return t.CalcTileCoordinate(c, level)
}

func (t *TileCoord) WGS84ToWebMercatorTile(c Coordinate, level int) TileCoordinate {
	mercator := c.WGS84ToWebMercator()
	var resp TileCoordinate
	resp.X = int((mercator.X - t.OriX) / t.Resolution(level) / float64(t.TileSize))
	resp.Y = int(math.Abs(t.OriY-mercator.Y) / t.Resolution(level) / float64(t.TileSize))
	resp.Level = level
	return resp
}

func (t *TileCoord) WebMercatorTileBound(coords []Coordinate, level int) TileCoordinateBound {
	bound := make([]TileCoordinate, 0, 2)
	wg := errgroup.Group{}
	for _, coord := range coords {
		coordCopy := coord
		wg.Go(func() error {
			cd := t.WebMercatorTile(coordCopy, level)
			bound = append(bound, TileCoordinate{
				X: cd.X,
				Y: cd.Y,
			})
			return nil
		})
	}
	wg.Wait()
	return bound
}

func (t *TileCoord) WGS84TileBound(level int, coords []Coordinate) TileCoordinateBound {
	bound := make([]TileCoordinate, 0, 2)
	wg := errgroup.Group{}
	for _, coord := range coords {
		coordCopy := coord
		wg.Go(func() error {
			cd := t.CalcTileCoordinate(coordCopy, level)
			bound = append(bound, TileCoordinate{
				X:     cd.X,
				Y:     cd.Y,
				Level: level,
			})
			return nil
		})
	}
	wg.Wait()
	return bound
}

func (t *TileCoord) WGS84ToWebMercatorTileBound(level int, coords []Coordinate) TileCoordinateBound {
	bound := make([]TileCoordinate, 0, 2)
	wg := errgroup.Group{}
	for _, coord := range coords {
		coordCopy := coord
		wg.Go(func() error {
			cd := t.WGS84ToWebMercatorTile(coordCopy, level)
			bound = append(bound, TileCoordinate{
				X:     cd.X,
				Y:     cd.Y,
				Level: level,
			})
			return nil
		})
	}
	wg.Wait()
	return bound
}

func (bound TileCoordinateBound) Scope() TileCoordinateScope {
	if len(bound) < 2 {
		return TileCoordinateScope{}
	}
	x1, y1 := bound[0].X, bound[0].Y
	x2, y2 := bound[1].X, bound[1].Y

	minX := getMin(x1, x2)
	maxX := getMax(x1, x2)

	maxY := getMax(y1, y2)
	minY := getMin(y2, y1)
	return TileCoordinateScope{
		MinX:  minX,
		MaX:   maxX,
		MinY:  minY,
		MaY:   maxY,
		Level: bound[0].Level,
	}
}

func (bound TileCoordinateBound) Expand(fn func(x, y, level int)) bool {
	if len(bound) < 2 {
		return false
	}
	level := bound[0].Level
	x1, y1 := bound[0].X, bound[0].Y
	x2, y2 := bound[1].X, bound[1].Y

	minX := getMin(x1, x2)
	maxX := getMax(x1, x2)

	maxY := getMax(y1, y2)
	minY := getMin(y2, y1)
	for y := minY; y <= maxY; y++ {
		yCopy := y
		for x := minX; x <= maxX; x++ {
			xCopy := x
			fn(xCopy, yCopy, level)
		}
	}

	return true
}

func getMin(x1, x2 int) int {
	if x1 > x2 {
		return x2
	}
	return x1
}

func getMax(x1, x2 int) int {
	if x1 > x2 {
		return x1
	}
	return x2
}
