package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"strconv"

	"github.com/usedbytes/osgrid"
	"github.com/usedbytes/osgrid/terrain50"
)

var gridRef string
var dataDir string
var radius uint
var outputFile string
var scadFile string
var horizontalScale string
var verticalScale string

func init() {
	const (
		// Snowdon summit
		defaultGridRef = "SH 60986 54375"
		usageGridRef   = "Centre point grid reference (Default is Snowdon summit)"

		usageDataDir   = "Database directory (should contain 'data' folder)"

		defaultRadius = 5 * osgrid.Kilometre
		usageRadius   = "Radius of map (metres)"

		defaultOutput = "-"
		usageOutput = "Output data file ('-' for stdout)"

		defaultSCAD = ""
		usageSCAD = "OpenSCAD output file (output file must be specified)"

		defaultHorizontalScale = "1:100000"
		usageHorizontalScale = "Horizontal scale (only affects OpenSCAD output)"

		defaultVerticalScale = "1:10000"
		usageVerticalScale = "Vertical scale (only affects OpenSCAD output)"
	)

	flag.StringVar(&gridRef, "grid", defaultGridRef, usageGridRef)
	flag.StringVar(&gridRef, "g", defaultGridRef, usageGridRef + " (shorthand)")

	flag.StringVar(&dataDir, "database", "", usageDataDir)
	flag.StringVar(&dataDir, "d", "", usageDataDir + " (shorthand)")

	flag.UintVar(&radius, "radius", defaultRadius, usageRadius)
	flag.UintVar(&radius, "r", defaultRadius, usageRadius + " (shorthand)")

	flag.StringVar(&outputFile, "output", defaultOutput, usageOutput)
	flag.StringVar(&outputFile, "o", defaultOutput, usageOutput + " (shorthand)")

	flag.StringVar(&scadFile, "scad", defaultSCAD, usageSCAD)
	flag.StringVar(&scadFile, "s", defaultSCAD, usageSCAD + " (shorthand)")

	flag.StringVar(&horizontalScale, "xyscale", defaultHorizontalScale, usageHorizontalScale)
	flag.StringVar(&horizontalScale, "x", defaultHorizontalScale, usageHorizontalScale + " (shorthand)")

	flag.StringVar(&verticalScale, "zscale", defaultVerticalScale, usageVerticalScale)
	flag.StringVar(&verticalScale, "z", defaultVerticalScale, usageVerticalScale + " (shorthand)")
}

func main() {
	flag.Parse()

	if len(dataDir) == 0 {
		log.Fatal("Database directory is required")
	}

	d, err := terrain50.OpenDatabase(dataDir, 10 * osgrid.Kilometre)
	if err != nil {
		log.Fatal(err)
	}

	dataOut := os.Stdout
	if outputFile != "-" {
		dataOut, err = os.Create(outputFile)
		if err != nil {
			log.Fatal(err)
		}
		defer dataOut.Close()
	}

	var scadOut *os.File
	if len(scadFile) > 0 {
		if len(outputFile) == 0 || outputFile == "-" {
			log.Fatal("OpenSCAD output (--scad) requires a data file output (--output)")
		}

		scadOut, err = os.Create(scadFile)
		if err != nil {
			log.Fatal(err)
		}
		defer scadOut.Close()
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

	maxElevation := float32(0.0)

	for north := osgrid.Distance(0); north < osgrid.Distance(radius * 2); north += d.Precision() {
		for east := osgrid.Distance(0); east < osgrid.Distance(radius * 2); east += d.Precision() {
			ref, err := bottomLeft.Add(east, north)
			if err != nil {
				log.Fatal(err)
			}

			val, err := d.GetData(ref)
			if err != nil {
				log.Fatal(err)
			}

			if val > maxElevation {
				maxElevation = val
			}

			fmt.Fprintf(dataOut, "%d ", int(val * 10))
		}
		fmt.Fprintln(dataOut, "")
	}

	log.Println("Max elevation:", maxElevation)

	if scadOut != nil {
		fields := strings.Split(horizontalScale, ":")
		if len(fields) != 2 {
			log.Fatalf("Invalid horizontal scale: %s\n", horizontalScale)
		}
		hNum, err := strconv.Atoi(fields[0])
		if err != nil {
			log.Fatalf("Invalid horizontal numerator: %s\n", fields[0])
		}
		hDen, err := strconv.Atoi(fields[1])
		if err != nil {
			log.Fatalf("Invalid horizontal denominator: %s\n", fields[0])
		}

		fields = strings.Split(verticalScale, ":")
		if len(fields) != 2 {
			log.Fatalf("Invalid vertical scale: %s\n", verticalScale)
		}
		vNum, err := strconv.Atoi(fields[0])
		if err != nil {
			log.Fatalf("Invalid vertical numerator: %s\n", fields[0])
		}
		vDen, err := strconv.Atoi(fields[1])
		if err != nil {
			log.Fatalf("Invalid vertical denominator: %s\n", fields[0])
		}

		// OpenSCAD uses millimetres, so multiply by 1000
		xSize := float32(radius * 2 * 1000) * float32(hNum) / float32(hDen)
		zSize := maxElevation * 1000 * float32(vNum) / float32(vDen)

		fmt.Fprintf(scadOut, "resize([%f, %f, %f]) surface(\"%s\", center = true);\n",
				xSize, xSize, zSize, outputFile)
	}
}
