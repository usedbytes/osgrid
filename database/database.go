package database

import (
	"image"
	"github.com/usedbytes/osgrid"
)

type Tile interface {
	Width() osgrid.Distance
	Height() osgrid.Distance
	Precision() osgrid.Distance

	BottomLeft() (osgrid.GridRef)
	String() string
}

type Float64Tile interface {
	GetFloat64(osgrid.GridRef) (float64, error)
}

type RasterTile interface {
	GetImage() image.Image
}

type Database interface {
	GetTile(osgrid.GridRef) (Tile, error)
	Precision() osgrid.Distance
}

type Float64Database interface {
	GetFloat64(osgrid.GridRef) (float64, error)
}

type RasterDatabase interface {
	GetImage() image.Image
}
