package osdata

import (
	"image"

	"github.com/usedbytes/osgrid"
)

type Tile interface {
	Width() osgrid.Distance
	Height() osgrid.Distance
	Precision() osgrid.Distance

	BottomLeft() osgrid.GridRef
	String() string
}

type Database interface {
	GetTile(osgrid.GridRef) (Tile, error)
	Precision() osgrid.Distance
}

type Float64Tile interface {
	Tile
	GetFloat64(osgrid.GridRef) (float64, error)
}

type Float64Database interface {
	Database
	GetFloat64(osgrid.GridRef) (float64, error)
	GetFloat64Tile(osgrid.GridRef) (Float64Tile, error)
}

type ImageTile interface {
	Tile
	GetImage() image.Image
	GetPixelCoord(ref osgrid.GridRef) (int, int, error)
}

type ImageDatabase interface {
	Database
	GetImageTile(osgrid.GridRef) (ImageTile, error)
}
