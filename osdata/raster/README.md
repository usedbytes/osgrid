# `raster`

Much of the data under
[OS OpenData](https://osdatahub.os.uk/downloads/open).
is available in a "raster" format, with image files representing tiles.

One such dataset is the 
[OS VectorMap District](https://osdatahub.os.uk/downloads/open/VectorMapDistrict)
set, which this package was developed against, though it may work with other
raster datasets, with no or minimal modification - I haven't tested that.

This package provides a way to work with a raster dataset, querying images
and pixel locations corresponding to grid references.

The `raster.Database` object represents the data set. To use it, simply
download the _Terrain 50_ "ASCII Grid" dataset, and extract it to a location. Use
this location to open the database. The path should be to the directory _which
contains the `data` directory_:

```
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
```

A tile cache is used (with 16 entries by default), storing the parsed data for
the 16 most-recently-used tiles so that queries which are geographically close
to each other are fast, and to ensure memory usage doesn't grow unbounded.

The `GenerateSurface()` function in `lib/geometry` provides the functionality to
query elevation data for a rectangular region.

## Known Issues

The `golang.org/x/image/tiff` package which is used for decoding the TIFF images
has an issue where it throws "unexpected EOF" for some images. For example tile
`NO08` exhibits this problem.

I've posted a patch which appears to fix it:
https://github.com/golang/go/issues/30827#issuecomment-774469551

I probably should either make a vendored version of that module and pin it in
this package, or try out a different TIFF decoder package.
