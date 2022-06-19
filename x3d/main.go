package main

import (
	"encoding/xml"
	"flag"
	"log"
	"math"
	"os"
	"strings"
	"strconv"

	"github.com/usedbytes/osgrid"
	"github.com/usedbytes/osgrid/osdata/terrain50"
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

	bottomLeft, err := centre.Add(osgrid.Distance(-radius), osgrid.Distance(-radius))
	if err != nil {
		log.Fatal(err)
	}

	// Horizontal in the source data
	inStep := d.Precision() * osgrid.Distance(decimate)

	// Horizontal step in the X3D file
	outStep := float64(inStep) * hScale

	mesh := &Mesh{
		Width:  (int(radius * 2) / int(inStep)) + 1,
		Height: (int(radius * 2) / int(inStep)) + 1,
	}

	baseMesh := &Mesh{
		Width:  (int(radius * 2) / int(inStep)) + 1,
		Height: (int(radius * 2) / int(inStep)) + 1,
	}

	texSStep := 1.0 / float64(mesh.Width-1)
	texTStep := 1.0 / float64(mesh.Height-1)

	texCoord := MFVec2f{}
	baseTexCoord := MFVec2f{}
	wallThickness := 5.0

	for y := 0; y < mesh.Height; y++ {
		for x := 0; x < mesh.Width; x++ {
			ref, err := bottomLeft.Add(osgrid.Distance(x) * inStep, osgrid.Distance(y) * inStep)
			if err != nil {
				log.Fatal(err)
			}

			val, err := d.GetData(ref)
			if err != nil {
				log.Fatal(err)
			}

			z := (float64(val) + float64(zOffset)) * vScale

			mesh.Points = append(mesh.Points, [3]float64{
				float64(x) * outStep,
				float64(y) * outStep,
				z,
			})
			texCoord = append(texCoord, [2]float64{float64(x) * texSStep, float64(y) * texTStep})

			if (float64(x) * outStep >= wallThickness) &&
			   (float64(x) * outStep <= (float64(mesh.Width) * outStep) - wallThickness) &&
			   (float64(y) * outStep >= wallThickness) &&
			   (float64(y) * outStep <= (float64(mesh.Height) * outStep) - wallThickness) {
				baseMesh.Points = append(baseMesh.Points, [3]float64{
					float64(x) * outStep,
					float64(y) * outStep,
					math.Max(0, float64(z) - wallThickness),
				})
			} else {
				baseMesh.Points = append(baseMesh.Points, [3]float64{
					float64(x) * outStep,
					float64(y) * outStep,
					0,
				})
			}

			//baseTexCoord = append(baseTexCoord, [2]float64{1.0 - (float64(x) * texSStep), float64(y) * texTStep})
			baseTexCoord = append(baseTexCoord, [2]float64{1.1, 1.1})
		}
	}

	x3d := &X3D{
		Version: 3.2,
		Profile: "Interchange",
		Scene: &Scene{
			Shape: &Shape{
				Appearance: &Appearance{
					ImageTexture: &ImageTexture{
						URL: "texture.png",
						TextureProperties: &TextureProperties{
							BoundaryModeS: "CLAMP_TO_BOUNDARY",
							BoundaryModeT: "CLAMP_TO_BOUNDARY",
						},
					},
				},
				IndexedFaceSet: &IndexedFaceSet{
					CCW: true,
					CoordIndex: mesh.Triangles(true),
					Coordinate: &Coordinate{
						Point: mesh.Points,
					},
					TextureCoordinate: &TextureCoordinate{
						Point: texCoord,
					},
				},
			},
		},
	}


	// Append the base mesh
	coord := x3d.Scene.Shape.IndexedFaceSet.Coordinate
	ifs := x3d.Scene.Shape.IndexedFaceSet
	firstBaseIdx := len(coord.Point)

	coord.Point = append(coord.Point, baseMesh.Points...)
	ifs.TextureCoordinate.Point = append(ifs.TextureCoordinate.Point, baseTexCoord...)

	// Note the winding order is set to clockwise, as we want the normals
	// to be reversed
	baseIndices := baseMesh.Triangles(false)
	for i := range baseIndices {
		val := baseIndices[i]
		if val != -1 {
			ifs.CoordIndex = append(ifs.CoordIndex, int32(firstBaseIdx) + val)
		} else {
			ifs.CoordIndex = append(ifs.CoordIndex, -1)
		}
	}

	// For each side, build a list of indices for the top and
	// bottom row of a triangle strip
	topIdx := make(MFInt32, mesh.Width)
	bottomIdx := make(MFInt32, mesh.Width)

	// South face
	// Base is SW to SE, top is the first row of points
	firstIdx := 0
	for i := 0; i < mesh.Width; i++ {
		topIdx[i] = int32(firstIdx + i)
		bottomIdx[i] = topIdx[i] + int32(firstBaseIdx)
	}
	ifs.CoordIndex = append(ifs.CoordIndex, MakeTriangleStrip(bottomIdx, topIdx, true)...)

	// North face
	// Base is NE to NW, top is the last row of points reversed
	firstIdx = mesh.Width*mesh.Height - 1
	for i := 0; i < mesh.Width; i++ {
		topIdx[i] = int32(firstIdx - i)
		bottomIdx[i] = topIdx[i] + int32(firstBaseIdx)
	}
	ifs.CoordIndex = append(ifs.CoordIndex, MakeTriangleStrip(bottomIdx, topIdx, true)...)

	// West face
	// Base is NW to SW, top is the first column of points reversed
	firstIdx = mesh.Width * (mesh.Height-1)
	for i := 0; i < mesh.Height; i++ {
		topIdx[i] = int32(firstIdx - i * mesh.Width)
		bottomIdx[i] = topIdx[i] + int32(firstBaseIdx)
	}
	ifs.CoordIndex = append(ifs.CoordIndex, MakeTriangleStrip(bottomIdx, topIdx, true)...)

	// East face
	// Base is SE to NE, top is the last column of points
	firstIdx = mesh.Width-1
	for i := 0; i < mesh.Height; i++ {
		topIdx[i] = int32(firstIdx + i * mesh.Width)
		bottomIdx[i] = topIdx[i] + int32(firstBaseIdx)
	}
	ifs.CoordIndex = append(ifs.CoordIndex, MakeTriangleStrip(bottomIdx, topIdx, true)...)

	enc := xml.NewEncoder(dataOut)
	enc.Indent("", "\t")

	procInst := xml.ProcInst{
		Target: "xml",
		Inst:   []byte("version=\"1.0\" encoding=\"UTF-8\""),
	}
	err = enc.EncodeToken(procInst)
	if err != nil {
		log.Println("ERROR:", err)
		return
	}

	dir := xml.Directive("DOCTYPE X3D PUBLIC \"ISO//Web3D//DTD X3D 3.2//EN\" \"https://www.web3d.org/specifications/x3d-3.2.dtd\"")
	err = enc.EncodeToken(dir)
	if err != nil {
		log.Println("ERROR:", err)
		return
	}

	err = enc.Encode(x3d)
	if err != nil {
		log.Println("ERROR:", err)
	}
}
