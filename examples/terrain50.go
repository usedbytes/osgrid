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
