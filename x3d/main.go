package main

import (
	"encoding/xml"
	"flag"
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
var decimate uint
var zOffset int

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

		// FIXME: This can slightly mess up the physical size because the
		// scaling doesn't take into account if the expected size isn't an
		// exact multiple of 'M'
		defaultDecimate = 1
		usageDecimate = "Decimate (only use every M'th sample) to reduce number of points"

		defaultZOffset = 0
		usageZOffset = "Amount (in metres) to add or subtract from all values"
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

	flag.UintVar(&decimate, "deciMate", defaultDecimate, usageDecimate)
	flag.UintVar(&decimate, "M", defaultDecimate, usageDecimate + " (shorthand)")

	flag.IntVar(&zOffset, "zOffset", defaultZOffset, usageZOffset)
	flag.IntVar(&zOffset, "Z", defaultZOffset, usageZOffset + " (shorthand)")
}

func main() {
	var err error

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

	centre, err := osgrid.ParseGridRef(gridRef)
	if err != nil {
		log.Fatal(err)
	}

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

	hScale := 1000 * float64(hNum) / float64(hDen)

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

	vScale := 1000 * float64(vNum) / float64(vDen)

	log.Printf("Centre: %s\n", centre.String())
	log.Printf("Radius: %d m\n", radius)
	log.Printf("Output: %s\n", outputFile)
	log.Printf("HScale: %f\n", hScale)
	log.Printf("VScale: %f\n", vScale)

	topLeft, err := centre.Add(osgrid.Distance(-radius), osgrid.Distance(radius))
	if err != nil {
		log.Fatal(err)
	}

	mesh := &Mesh{ }

	// In x3d coordinate system:
	// 'x': west -> east
	// 'y': elevation
	// 'z': north -> south

	for south := osgrid.Distance(0); south < osgrid.Distance(radius * 2); south += d.Precision() * osgrid.Distance(decimate) {
		for east := osgrid.Distance(0); east < osgrid.Distance(radius * 2); east += d.Precision() * osgrid.Distance(decimate) {
			ref, err := topLeft.Add(east, -south)
			if err != nil {
				log.Fatal(err)
			}

			val, err := d.GetData(ref)
			if err != nil {
				log.Fatal(err)
			}

			mesh.Points = append(mesh.Points, [3]float64{
				float64(east) * hScale,
				(float64(val) + float64(zOffset)) * vScale,
				float64(south) * hScale,
			})
		}

		// HAX
		mesh.Width++
		mesh.Height++
	}

	x3d := &X3D{
		Version: 3.2,
		Profile: "interchange",
		Scene: &Scene{
			Shape: &Shape{
				IndexedFaceSet: &IndexedFaceSet{
					CCW: true,
					CoordIndex: mesh.Triangles(),
					Coordinate: &Coordinate{
						Point: mesh.Points,
					},
				},
			},
		},
	}

	coord := x3d.Scene.Shape.IndexedFaceSet.Coordinate
	ifs := x3d.Scene.Shape.IndexedFaceSet

	// Add 4 more faces:
	// Bottom: (0, 0, 0), (eastMax, 0, 0), (eastMax, 0, southMax), (0, 0, southMax)
	bottomPoints := MFVec3f{
		{0, 0, 0},
		{float64(radius * 2) * hScale, 0, 0},
		{float64(radius * 2) * hScale, 0, float64(radius * 2) * hScale},
		{0, 0, float64(radius * 2) * hScale},
	}
	idx := int32(len(coord.Point))
	coord.Point = append(coord.Point, bottomPoints...)
	ifs.CoordIndex = append(ifs.CoordIndex, MFInt32{ idx, idx + 1, idx + 2, idx + 3, -1 }...)

	// North:  (0, 0, 0), [first row of points], (eastMax, 0, 0)
	//         idx, 0:width, idx+1
	northFace := MFInt32{ idx }
	for i := int32(0); i < mesh.Width; i++ {
		northFace = append(northFace, i)
	}
	northFace = append(northFace, MFInt32{ idx+1, -1 }...)
	ifs.CoordIndex = append(ifs.CoordIndex, northFace...)

	// West:   (0, 0, 0), (0, 0, southMax), [first col of points reversed]
	//         idx, idx+3, width*(height-1):-width:0
	westFace := MFInt32{ idx, idx + 3 }
	for i := mesh.Width*(mesh.Height-1); i >= 0; i-=mesh.Width {
		westFace = append(westFace, i)
	}
	westFace = append(westFace, MFInt32{ -1 }...)
	ifs.CoordIndex = append(ifs.CoordIndex, westFace...)

	// South:  (0, 0, southMax), (eastMax, 0, southMax), [last row of points reversed]
	//         idx+3, idx+2, (width*height)-1:-1:width*(height-1)
	southFace := MFInt32{ idx+3, idx+2 }
	for i := mesh.Width*mesh.Height-1; i >= mesh.Width*(mesh.Height-1); i-=1 {
		southFace = append(southFace, i)
	}
	southFace = append(southFace, MFInt32{ -1 }...)
	ifs.CoordIndex = append(ifs.CoordIndex, southFace...)

	// East:   (eastMax, 0, southMax), (eastMax, 0, 0), [last col of points]
	//         idx+2, idx+1, width-1:width:(width*height)-1
	eastFace := MFInt32{ idx+2, idx+1 }
	for i := mesh.Width-1; i < mesh.Width*mesh.Height; i+=mesh.Width {
		eastFace = append(eastFace, i)
	}
	eastFace = append(eastFace, MFInt32{ -1 }...)
	ifs.CoordIndex = append(ifs.CoordIndex, eastFace...)

	enc := xml.NewEncoder(dataOut)
	enc.Indent("", "\t")

	err = enc.Encode(x3d)
	if err != nil {
		log.Println("ERROR:", err)
	}
}
