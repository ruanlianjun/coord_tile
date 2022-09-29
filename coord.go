package coordtile

type CoordTransformI interface {
	//LL->web mercator
	WGS84ToWebMercator(coord Coordinate) Coordinate
	//web mercator->LL
	WebMercatorToWGS84(coord Coordinate) Coordinate
}

type TileCoordCalc interface {
}
