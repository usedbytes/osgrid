package main

import (
	"flag"
	"image"
	"image/png"
	"image/draw"
	"log"
	"math"
	"os"

	"github.com/nfnt/resize"
	"github.com/usedbytes/osgrid"
	"github.com/usedbytes/osgrid/database"
	"github.com/usedbytes/osgrid/raster"
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

	d, err := raster.OpenDatabase(dataDir, 10 * osgrid.Kilometre)
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

	dbtile, err := d.GetTile(bottomLeft)
	if err != nil {
		log.Fatal(err)
	}

	// FIXME:
	scale := 2.5

	pixelWidth := int(math.Round(float64((radius * 2)) / scale))
	downScale := 10
	//pixelWidth /= downScale
	canvas := image.NewRGBA(image.Rect(0, 0, pixelWidth, pixelWidth))

	drawnMinY := pixelWidth
	drawnMaxX := 0

	rowStart := bottomLeft

	for drawnMinY > 0 {
		dy := 0

		log.Println("top rowStart:", rowStart)
		coord := rowStart

		for drawnMaxX < pixelWidth {
			log.Println("drawnMaxX:", drawnMaxX, "pixelWidth:", pixelWidth)
			log.Println("drawnMinY:", drawnMinY)
			log.Println("coord:", coord)

			dbtile, err := d.GetTile(coord)
			if err != nil {
				log.Fatal("GetTile: ", err)
			}

			tile := dbtile.(database.ImageTile)

			log.Println("tile:", tile.BottomLeft())

			img := tile.GetImage()

			// Pixel coordinate of the bottom left of this patch
			minX, maxY, err := tile.GetPixelCoord(coord)
			log.Println("minX, maxY:", minX, maxY)

			// How far right can we go within this tile?
			// Assume we need the whole width to start with
			maxX := img.Bounds().Dx()
			// Then check if the right hand edge is actually within this tile
			tileBottomRight, _ := tile.BottomLeft().Add(tile.Width(), 0)
			if tileBottomRight.Tile() == bottomRight.Tile() && tileBottomRight.TileEasting() > bottomRight.TileEasting() {
				// Find the pixel coordinate of the right edge.
				eastDistance := bottomRight.TileEasting() - tile.BottomLeft().TileEasting()
				regionRightEdge, _ := tile.BottomLeft().Add(eastDistance, 0)

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
			tileTopRight, _ := tile.BottomLeft().Add(0, tile.Height())
			if tileTopRight.Tile() == topRight.Tile() && tileTopRight.TileNorthing() > topRight.TileNorthing() {
				// Find the pixel coordinate of the top edge.
				northDistance := topRight.TileNorthing() - tile.BottomLeft().TileNorthing()
				regionTopEdge, _ := tile.BottomLeft().Add(0, northDistance)

				_, minY, err = tile.GetPixelCoord(regionTopEdge)
				if err != nil {
					log.Fatalln("top right should be in tile:", err)
				}
			}
			log.Println("minY:", minY)

			//minX /= downScale
			//maxX /= downScale
			//minY /= downScale
			//maxY /= downScale
			//img := resize.Resize(img.Bounds().Dx() / downScale, img.Bounds.Dy() / downScale, img, resize.Lanczos3)


			// Copy Rect(minX, minY, maxX, maxY)
			dp := image.Pt(drawnMaxX, drawnMinY - (maxY - minY))
			sr := image.Rect(minX, minY, maxX, maxY)
			dr := image.Rectangle{dp, dp.Add(sr.Size())}
			draw.Draw(canvas, dr, img, sr.Min, draw.Src)

			log.Println("copy:", sr, "->", dr)

			drawnMaxX += sr.Size().X
			dy = sr.Size().Y

			coord, err = tile.BottomLeft().Add(tile.Width(), coord.TileNorthing() - tile.BottomLeft().TileNorthing())
			if err != nil {
				log.Fatal(err)
			}
		}

		// Start the next row
		drawnMaxX = 0
		drawnMinY -= dy

		//coord, err = tile.BottomLeft().Add(tile.Width(), coord.TileNorthing() - tile.BottomLeft().TileNorthing())
		rowStart, err = rowStart.Add(0, dbtile.Height())
		if err != nil {
			log.Fatal(err)
		}
		aligned := rowStart.Align(dbtile.Height())
		rowStart, _ = rowStart.Add(0, -(rowStart.TileNorthing() - aligned.TileNorthing()))
		//rowStart, _ = rowStart.Align(tile.Height()).Add(bottomLeft.TileEasting() - rowStart.TileEasting(), 0)
		log.Println("New row", drawnMinY)
		log.Println("New rowStart", rowStart)
	}

	dataOut, err := os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer dataOut.Close()

	small := resize.Resize(uint(pixelWidth / downScale), uint(pixelWidth / downScale), canvas, resize.Lanczos3)

	//png.Encode(dataOut, canvas)
	png.Encode(dataOut, small)
}

