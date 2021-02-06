package main

import (
	"flag"
	"image"
	"image/png"
	"image/draw"
	"log"
	"math"
	"os"

	"github.com/usedbytes/osgrid"
	"github.com/usedbytes/osgrid/vector"
)

var gridRef string
var dataDir string
var radius uint
var outputFile string

func init() {
	const (
		// Snowdon summit
		defaultGridRef = "SH 60986 54375"
		usageGridRef   = "Centre point grid reference (Default is Snowdon summit)"

		usageDataDir   = "Database directory (should contain 'data' folder)"

		defaultRadius = 5 * osgrid.Kilometre
		usageRadius   = "Radius of map (metres)"

		defaultOutput = ""
		usageOutput = "Output data file ('-' for stdout)"
	)

	flag.StringVar(&gridRef, "grid", defaultGridRef, usageGridRef)
	flag.StringVar(&gridRef, "g", defaultGridRef, usageGridRef + " (shorthand)")

	flag.StringVar(&dataDir, "database", "", usageDataDir)
	flag.StringVar(&dataDir, "d", "", usageDataDir + " (shorthand)")

	flag.UintVar(&radius, "radius", defaultRadius, usageRadius)
	flag.UintVar(&radius, "r", defaultRadius, usageRadius + " (shorthand)")

	flag.StringVar(&outputFile, "output", defaultOutput, usageOutput)
	flag.StringVar(&outputFile, "o", defaultOutput, usageOutput + " (shorthand)")
}

func main() {
	flag.Parse()

	if len(dataDir) == 0 {
		log.Fatal("Database directory is required")
	}

	if len(outputFile) == 0 {
		log.Fatal("Output file is required")
	}

	d, err := vector.OpenDatabase(dataDir, 10 * osgrid.Kilometre)
	if err != nil {
		log.Fatal(err)
	}

	centre, err := osgrid.ParseGridRef(gridRef)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Centre: %s\n", centre.String())
	log.Printf("Radius: %d m\n", radius)
	log.Printf("Output: %s\n", outputFile)

	bottomLeft, err := centre.Add(osgrid.Distance(-radius), osgrid.Distance(-radius))
	if err != nil {
		log.Fatal(err)
	}

	bottomRight, err := centre.Add(osgrid.Distance(radius), osgrid.Distance(-radius))
	if err != nil {
		log.Fatal(err)
	}

	topRight, err := centre.Add(osgrid.Distance(radius), osgrid.Distance(radius))
	if err != nil {
		log.Fatal(err)
	}

	tile, err := d.GetTile(bottomLeft)
	if err != nil {
		log.Fatal(err)
	}

	// FIXME:
	scale := 2.5

	pixelWidth := int(math.Round(float64((radius * 2)) / scale))
	canvas := image.NewRGBA(image.Rect(0, 0, pixelWidth, pixelWidth))

	drawnMinY := pixelWidth
	drawnMaxX := 0

	rowStart := bottomLeft
	coord := rowStart

	for drawnMinY > 0 {
		dy := 0

		for drawnMaxX < pixelWidth {
			log.Println("drawnMaxX:", drawnMaxX, "pixelWidth:", pixelWidth)
			log.Println("drawnMinY:", drawnMinY)
			tile, err = d.GetTile(coord)
			if err != nil {
				log.Fatal(err)
			}

			img := tile.Image()

			log.Println("coord:", coord, "tile:", tile.GridRef())

			// Pixel coordinate of the bottom left of this patch
			minX, maxY, err := tile.GetPixelCoord(coord)
			log.Println("minX, maxY:", minX, maxY)

			// How far right can we go within this tile?
			// Assume we need the whole width to start with
			maxX := img.Bounds().Dx()
			// Then check if the right hand edge is actually within this tile
			tileBottomRight, _ := tile.GridRef().Add(tile.Width(), 0)
			if tileBottomRight.Tile() == bottomRight.Tile() && tileBottomRight.TileEasting() > bottomRight.TileEasting() {
				// Find the pixel coordinate of the right edge.
				eastDistance := bottomRight.TileEasting() - tile.GridRef().TileEasting()
				regionRightEdge, _ := tile.GridRef().Add(eastDistance, 0)

				maxX, _, err = tile.GetPixelCoord(regionRightEdge)
				if err != nil {
					log.Fatalln("bottom right should be in tile:", err)
				}
			}
			log.Println("maxX:", maxX)

			// How far up can we go within this tile?
			// Assume we need the whole height to start with
			minY := 0
			// Then check if the top edge is actually within this tile
			tileTopRight, _ := tile.GridRef().Add(0, tile.Height())
			if tileTopRight.Tile() == topRight.Tile() && tileTopRight.TileNorthing() > topRight.TileNorthing() {
				// Find the pixel coordinate of the top edge.
				northDistance := topRight.TileNorthing() - tile.GridRef().TileNorthing()
				regionTopEdge, _ := tile.GridRef().Add(0, northDistance)

				_, minY, err = tile.GetPixelCoord(regionTopEdge)
				if err != nil {
					log.Fatalln("top right should be in tile:", err)
				}
			}
			log.Println("minY:", minY)


			// Copy Rect(minX, minY, maxX, maxY)
			dp := image.Pt(drawnMaxX, drawnMinY - (maxY - minY))
			sr := image.Rect(minX, minY, maxX, maxY)
			dr := image.Rectangle{dp, dp.Add(sr.Size())}
			draw.Draw(canvas, dr, img, sr.Min, draw.Src)

			log.Println("copy:", sr, "->", dr)

			drawnMaxX += sr.Size().X
			dy = sr.Size().Y

			coord, err = tile.GridRef().Add(tile.Width(), coord.TileNorthing() - tile.GridRef().TileNorthing())
			if err != nil {
				log.Fatal(err)
			}
		}

		// Start the next row
		drawnMaxX = 0
		drawnMinY -= dy

		rowStart, err = rowStart.Add(0, tile.Height())
		if err != nil {
			log.Fatal(err)
		}
	}

	dataOut, err := os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer dataOut.Close()

	png.Encode(dataOut, canvas)
}

