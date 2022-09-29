package coordtile

import (
	"fmt"
	"testing"
)

func TestTileCoord_GetTile(t1 *testing.T) {
	//15 X: 120.8657455444336, Y: 30.759538817987497 x: 27380, y: 13434,
	coord := NewTileCoord(TileProjWebMercator)
	//13 120.805664, 30.798474 x: 6844, y: 3358,
	tile := coord.WGS84ToWebMercatorTile(Coordinate{
		X: 120.81098556518555,
		Y: 30.803634881295125,
	}, 15)
	fmt.Println(tile)

	tile = coord.WGS84ToWebMercatorTile(Coordinate{
		X: 120.805664,
		Y: 30.798474,
	}, 13)
	fmt.Println(tile)

	tile = coord.WGS84ToWebMercatorTile(Coordinate{
		X: 120.8657455444336,
		Y: 30.759538817987497,
	}, 13)
	fmt.Println(tile)
}

func TestTileCoord_WGS84ToWebMercatorTileBounds(t1 *testing.T) {
	coord := NewTileCoord(TileProjWebMercator)

	bound := coord.WGS84ToWebMercatorTileBound(13, []Coordinate{
		{
			X: 120.805664, Y: 30.798474,
		},
		{
			X: 120.816650, Y: 30.807911,
		},
	})
	fmt.Println("bound:", bound)
	//x:6844 y:3358 level:13
	//x:6845 y:3358 level:13
	bound.Expand(func(x, y, level int) {
		fmt.Printf("---------x:%d y:%d level:%d\n", x, y, level)
	})

}
