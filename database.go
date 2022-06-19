package osgrid

import (
	"image"
)

type Tile interface {
	Width() Distance
	Height() Distance
	Precision() Distance

	BottomLeft() (GridRef)
	String() string
}

type Float64Tile interface {
	Tile
	GetFloat64(GridRef) (float64, error)
}

type ImageTile interface {
	Tile
	GetImage() image.Image
	GetPixelCoord(ref GridRef) (int, int, error)
}

type Database interface {
	GetTile(GridRef) (Tile, error)
	Precision() Distance
}

type Float64Database interface {
	Database
	GetFloat64(GridRef) (float64, error)
	GetFloat64Tile(GridRef) (Float64Tile, error)
}

type ImageDatabase interface {
	Database
	GetImageTile(GridRef) (ImageTile, error)
}
