# `terrain50`

One of the datasets available under
[OS OpenData](https://osdatahub.os.uk/downloads/open).
is
[_Terrain 50_](https://osdatahub.os.uk/downloads/open/Terrain50),
which provides elevation data for the whole of the UK, with 0.1 m vertical and
50 m horizontal resolution.

This package provides a way to work with this dataset, giving easy access to
elevation data by grid-reference.

The `terrain50.Database` object represents the data set. To use it, simply
download the _Terrain 50_ "ASCII Grid" dataset, and extract it to a location. Use
this location to open the database. The path should be to the directory _which
contains the `data` directory_:

```
package main

import (
	"fmt"
	"os"

	"github.com/usedbytes/osgrid"
	"github.com/usedbytes/osgrid/osdata/terrain50"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Provide path to raster data as only argument")
		return
	}

	db, err := terrain50.OpenDatabase(os.Args[1], 10 * osgrid.Kilometre)
	if err != nil {
		panic(err)
	}

	summit, _ := osgrid.ParseGridRef("SH 60986 54375")
	elevation, _ := db.GetFloat64(summit)
	fmt.Printf("Snowdon's summit is at %f m\n", elevation)
}
```

A tile cache is used (with 16 entries by default), storing the parsed data for
the 16 most-recently-used tiles so that queries which are geographically close
to each other are fast, and to ensure memory usage doesn't grow unbounded.

The `GenerateSurface()` function in `lib/geometry` provides the functionality to
query elevation data for a rectangular region.
