package main

import (
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"image/png"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/usedbytes/osgrid"
	"github.com/usedbytes/osgrid/lib/geometry"
	"github.com/usedbytes/osgrid/lib/texture"
	"github.com/usedbytes/osgrid/lib/x3d"
	"github.com/usedbytes/osgrid/osdata"
	"github.com/usedbytes/osgrid/osdata/raster"
	"github.com/usedbytes/osgrid/osdata/terrain50"
)

type meshOutputOpts struct {
	textureFilename string
}

type meshConfig struct {
	elevationDB osdata.Float64Database
	rasterDB    osdata.ImageDatabase
	width       osgrid.Distance
	outFile     io.WriteCloser
	gridRef     osgrid.GridRef
	formatter   MeshFormatter
	meshOpts    []geometry.GenerateMeshOpt
	textureOpts []texture.GenerateTextureOpt
	outputOpts  meshOutputOpts
}

type MeshFormatter func(io.Writer, *geometry.Mesh, *meshOutputOpts) error

func writeMeshSTL(w io.Writer, m *geometry.Mesh, opts *meshOutputOpts) error {
	hdr := make([]byte, 80)
	copy(hdr, []byte("shape osmodel"))

	err := binary.Write(w, binary.LittleEndian, hdr)
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, uint32(len(m.Triangles)))
	if err != nil {
		return err
	}

	for _, t := range m.Triangles {
		err = binary.Write(w, binary.LittleEndian, [3]float32{0, 0, 0})
		if err != nil {
			return err
		}
		for _, idx := range t {
			vert := m.Vertices[idx]

			err = binary.Write(w, binary.LittleEndian, float32(vert[0]))
			if err != nil {
				return err
			}

			err = binary.Write(w, binary.LittleEndian, float32(vert[1]))
			if err != nil {
				return err
			}

			err = binary.Write(w, binary.LittleEndian, float32(vert[2]))
			if err != nil {
				return err
			}
		}

		err = binary.Write(w, binary.LittleEndian, uint16(0))
		if err != nil {
			return err
		}
	}

	return nil
}

func writeMeshX3D(w io.Writer, m *geometry.Mesh, opts *meshOutputOpts) error {
	x := &x3d.X3D{
		Version: 3.2,
		Profile: "Interchange",
		Scene: &x3d.Scene{
			Shape: &x3d.Shape{
				IndexedFaceSet: &x3d.IndexedFaceSet{
					CCW: true,
					Coordinate: &x3d.Coordinate{
						Point: m.Vertices,
					},
				},
			},
		},
	}

	indices := make(x3d.MFInt32, len(m.Triangles)*4)
	for i, t := range m.Triangles {
		indices[i*4+0] = int32(t[0])
		indices[i*4+1] = int32(t[1])
		indices[i*4+2] = int32(t[2])
		indices[i*4+3] = int32(-1)
	}

	x.Scene.Shape.IndexedFaceSet.CoordIndex = indices

	if len(m.TexCoords) == len(m.Vertices) {
		x.Scene.Shape.Appearance = &x3d.Appearance{
			ImageTexture: &x3d.ImageTexture{
				URL: opts.textureFilename,
				TextureProperties: &x3d.TextureProperties{
					BoundaryModeS: "CLAMP",
					BoundaryModeT: "CLAMP",
				},
			},
		}

		x.Scene.Shape.IndexedFaceSet.TextureCoordinate = &x3d.TextureCoordinate{
			Point: x3d.MFVec2f(m.TexCoords),
		}
	}

	enc := xml.NewEncoder(w)
	enc.Indent("", "\t")

	procInst := xml.ProcInst{
		Target: "xml",
		Inst:   []byte("version=\"1.0\" encoding=\"UTF-8\""),
	}
	err := enc.EncodeToken(procInst)
	if err != nil {
		return err
	}

	dir := xml.Directive("DOCTYPE X3D PUBLIC \"ISO//Web3D//DTD X3D 3.2//EN\" \"https://www.web3d.org/specifications/x3d-3.2.dtd\"")
	err = enc.EncodeToken(dir)
	if err != nil {
		return err
	}

	err = enc.Encode(x)
	if err != nil {
		return err
	}

	return nil
}

func writeMeshSCADPolyhedron(w io.Writer, m *geometry.Mesh, opts *meshOutputOpts) error {
	_, err := io.WriteString(w, "polyhedron(\n\t")
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, "points = [")
	if err != nil {
		return err
	}

	for i, v := range m.Vertices {
		end := ",\n\t"
		if i == len(m.Vertices)-1 {
			end = "],\n\t"
		}

		_, err = io.WriteString(w, fmt.Sprintf("[%f, %f, %f]%s", v[0], v[1], v[2], end))
		if err != nil {
			return err
		}
	}

	_, err = io.WriteString(w, "faces = [")
	if err != nil {
		return err
	}

	for i, t := range m.Triangles {
		end := ",\n\t"
		if i == len(m.Triangles)-1 {
			end = "]\n);"
		}

		_, err = io.WriteString(w, fmt.Sprintf("[%v, %v, %v]%s", t[0], t[1], t[2], end))
		if err != nil {
			return err
		}
	}

	return nil
}

func meshFormatterFromFormat(format string, c *cli.Context) (MeshFormatter, error) {
	switch format {
	case "scad":
		return func(w io.Writer, m *geometry.Mesh, opts *meshOutputOpts) error {
			return writeMeshSCADPolyhedron(w, m, opts)
		}, nil
	case "stl":
		return func(w io.Writer, m *geometry.Mesh, opts *meshOutputOpts) error { return writeMeshSTL(w, m, opts) }, nil
	case "x3d":
		return func(w io.Writer, m *geometry.Mesh, opts *meshOutputOpts) error { return writeMeshX3D(w, m, opts) }, nil
	default:
		return nil, fmt.Errorf("unknown format: %s", format)
	}
}

func parseScale(scale string) (float64, error) {
	fields := strings.Split(scale, ":")
	if len(fields) != 2 {
		return 0, fmt.Errorf("couldn't parse scale: '%s'\n", scale)
	}

	num, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, fmt.Errorf("couldn't parse scale numerator: '%s'\n", fields[0])
	}
	den, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, fmt.Errorf("couldn't parse scale denominator: '%s'\n", fields[1])
	}

	return float64(num) / float64(den), nil
}

func parseMeshArgs(c *cli.Context) (meshConfig, error) {
	var cfg meshConfig
	var err error

	var cleanup []func()
	success := false

	defer func() {
		if !success {
			for _, c := range cleanup {
				c()
			}
		}
	}()

	// elevation
	cfg.elevationDB, err = terrain50.OpenDatabase(c.String("elevation"), 10*osgrid.Kilometre)
	if err != nil {
		return meshConfig{}, fmt.Errorf("opening elevation database: %w", err)
	}

	// raster
	if c.Bool("texture") {
		if c.String("raster") == "" {
			return meshConfig{}, fmt.Errorf("--raster is required to generate textures")
		}

		cfg.rasterDB, err = raster.OpenDatabase(c.String("raster"), 10*osgrid.Kilometre)
		if err != nil {
			return meshConfig{}, fmt.Errorf("opening raster database: %w", err)
		}
	}

	// GRID_REFERENCE
	gridRef := snowdon
	if c.NArg() > 0 {
		gridRef = strings.Join(c.Args().Slice(), "")
	}

	cfg.gridRef, err = osgrid.ParseGridRef(gridRef)
	if err != nil {
		return meshConfig{}, fmt.Errorf("parsing GRID_REFERENCE: %w", err)
	}

	// hscale
	v, err := parseScale(c.String("hscale"))
	if err != nil {
		return meshConfig{}, fmt.Errorf("hscale: %w", err)
	}
	cfg.meshOpts = append(cfg.meshOpts, geometry.MeshHScaleOpt(v))

	// vscale
	v, err = parseScale(c.String("vscale"))
	if err != nil {
		return meshConfig{}, fmt.Errorf("vscale: %w", err)
	}
	cfg.meshOpts = append(cfg.meshOpts, geometry.MeshVScaleOpt(v))

	// width
	cfg.width = osgrid.Distance(c.Uint("width"))

	cfg.outFile, err = os.Create(c.String("outfile"))
	if err != nil {
		return meshConfig{}, fmt.Errorf("opening outfile: %w", err)
	}
	cleanup = append(cleanup, func() { cfg.outFile.Close() })

	format := path.Ext(c.String("outfile"))[1:]

	// format (overrides file extension if set)
	if c.IsSet("format") {
		format = c.String("format")
	}

	cfg.formatter, err = meshFormatterFromFormat(format, c)
	if err != nil {
		return meshConfig{}, err
	}

	if format == "stl" || format == "x3d" {
		cfg.meshOpts = append(cfg.meshOpts, geometry.MeshWindingOpt(true))
	}

	if c.Bool("texture") {
		if !c.IsSet("raster") {
			return meshConfig{}, fmt.Errorf("texture requested but raster database not provided")
		}

		if format != "x3d" {
			return meshConfig{}, fmt.Errorf("texturing only supported for X3D output")
		}

		outfile := c.String("outfile")
		cfg.outputOpts.textureFilename = outfile[:strings.LastIndex(outfile, ".")] + "_texture.png"
	}

	success = true

	return cfg, nil
}

func runMesh(c *cli.Context) error {
	cfg, err := parseMeshArgs(c)
	if err != nil {
		return err
	}
	defer cfg.outFile.Close()

	surface, err := geometry.GenerateSurface(cfg.elevationDB, cfg.gridRef, cfg.width, cfg.width)
	if err != nil {
		return fmt.Errorf("generating surface: %w", err)
	}

	if c.Bool("texture") {
		tex, err := texture.GenerateTexture(cfg.rasterDB, cfg.gridRef, cfg.width, cfg.width, cfg.textureOpts...)
		if err != nil {
			return fmt.Errorf("generating texture: %w", err)
		}

		f, err := os.Create(cfg.outputOpts.textureFilename)
		if err != nil {
			return err
		}
		defer f.Close()

		err = png.Encode(f, tex.Image)
		if err != nil {
			return err
		}

		texMap := texture.GenerateTextureMap(&tex, &surface)
		cfg.meshOpts = append(cfg.meshOpts, geometry.MeshTextureCoordsOpt(texMap.TexCoords))
	}

	mesh := geometry.GenerateMesh(&surface, cfg.meshOpts...)

	cfg.formatter(cfg.outFile, &mesh, &cfg.outputOpts)

	return nil
}

func hscaleFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:        "hscale",
		Aliases:     []string{"H"},
		Usage:       "Horizontal `SCALE` for output",
		Value:       "1:100",
		DefaultText: "1:100 - 10 mm per km, for output units of millimetres",
	}
}

func vscaleFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:        "vscale",
		Aliases:     []string{"V"},
		Usage:       "Vertical `SCALE` for output",
		Value:       "1:50",
		DefaultText: "1:50 - 20 mm per km, for output units of millimetres",
	}
}

func textureFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:    "texture",
		Aliases: []string{"t"},
		Usage:   "Generate and apply a texture using 'raster' data",
	}
}

var meshCmd cli.Command = cli.Command{
	Name: "mesh",
	Usage: "Generate a mesh from elevation data\n" +
		"\n" +
		"Default GRID_REFERENCE is Snowdon summit (" + snowdon + ")",
	ArgsUsage: "GRID_REFERENCE",
	Flags: []cli.Flag{
		elevationFlag(),
		formatsFlag([]string{"scad", "stl", "x3d"}),
		hscaleFlag(),
		outfileFlag(true),
		rasterFlag(false),
		textureFlag(),
		vscaleFlag(),
		widthFlag(),
	},
	Action: runMesh,
}
