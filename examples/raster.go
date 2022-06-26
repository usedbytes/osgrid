package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"

	"github.com/usedbytes/osgrid"
	"github.com/usedbytes/osgrid/osdata/raster"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Provide path to raster data as only argument")
		return
	}

	db, err := raster.OpenDatabase(os.Args[1], 10 * osgrid.Kilometre)
	if err != nil {
		panic(err)
	}

	summit, _ := osgrid.ParseGridRef("SH 60986 54375")

	tile, _ := db.GetImageTile(summit)

	x, y, _ := tile.GetPixelCoord(summit)

	out := image.NewRGBA(image.Rect(0, 0, 500, 500))

	draw.Draw(out, out.Bounds(), tile.GetImage(), image.Pt(x - 250, y - 250), draw.Over)

	f, err := os.Create("out.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	png.Encode(f, out)
}
