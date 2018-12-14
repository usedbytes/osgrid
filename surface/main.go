package main

import (
	"fmt"
	"log"

	"github.com/usedbytes/osgrid"
	"github.com/usedbytes/osgrid/terrain50"
)

func main() {
	d, err := terrain50.OpenDatabase("/aux/data/os_terrain", 10 * osgrid.Kilometre)
	if err != nil {
		log.Fatal(err)
	}

	centre, err := osgrid.ParseGridRef("ST 75337 64296")
	if err != nil {
		log.Fatal(err)
	}

	width, height := osgrid.Distance(15 * osgrid.Kilometre), osgrid.Distance(15 * osgrid.Kilometre)

	bottomLeft, err := centre.Add(-width / 2, -height / 2)
	if err != nil {
		log.Fatal(err)
	}

	max := float32(0.0)

	for north := osgrid.Distance(0); north < height; north += d.Precision() {
		for east := osgrid.Distance(0); east < width; east += d.Precision() {
			ref, err := bottomLeft.Add(east, north)
			if err != nil {
				log.Fatal(err)
			}

			val, err := d.GetData(ref)
			if err != nil {
				log.Fatal(err)
			}

			if val > max {
				max = val
			}

			fmt.Printf("%d ", int(val * 10))
		}
		fmt.Println("")
	}

	log.Println("Max:", max)

}
