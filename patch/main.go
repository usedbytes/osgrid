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

	minX, maxY, err := tile.GetPixelCoord(bottomLeft)
	blI := tile.Image()

	// Assume we need the whole width to start with
	maxX := blI.Bounds().Dx()
	// Then check if the right hand edge is actually within this tile
	tileBr, _ := tile.GridRef().Add(tile.Width(), 0)
	if tileBr.Tile() == tile.GridRef().Tile() && tileBr.TileEasting() <= bottomRight.TileEasting() {
		maxX, _, err = tile.GetPixelCoord(bottomRight)
		if err != nil {
			log.Fatalln("bottom right should be in tile", err)
		}
	}

	// Assume we need the whole height to start with
	minY := 0
	// Then check if the top edge is actually within this tile
	tileTr, _ := tile.GridRef().Add(0, tile.Height())
	if tileTr.Tile() == tile.GridRef().Tile() && tileTr.TileNorthing() <= topRight.TileEasting() {
		_, minY, err = tile.GetPixelCoord(topRight)
		if err != nil {
			log.Fatalln("top right should be in tile", err)
		}
	}

	// Copy Rect(minX, minY, maxX, maxY)

	dp := image.Pt(0, canvas.Bounds().Dy() - (maxY - minY))
	sr := image.Rect(minX, minY, maxX, maxY)
	dr := image.Rectangle{dp, sr.Size()}

	draw.Draw(canvas, dr, blI, sr.Min, draw.Src)

	dataOut, err := os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer dataOut.Close()

	png.Encode(dataOut, canvas)
}

