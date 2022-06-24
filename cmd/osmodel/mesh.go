package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/usedbytes/osgrid"
	"github.com/usedbytes/osgrid/lib/geometry"
	"github.com/usedbytes/osgrid/osdata"
	"github.com/usedbytes/osgrid/osdata/terrain50"
)

type MeshFormatter func(io.Writer, *geometry.Mesh) error

type meshConfig struct {
	elevationDB osdata.Float64Database
	width       osgrid.Distance
	outFile     io.WriteCloser
	gridRef     osgrid.GridRef
	formatter   MeshFormatter
	meshOpts    []geometry.GenerateMeshOpt
}

func writeMeshSCADPolyhedron(w io.Writer, m *geometry.Mesh) error {
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
		return func(w io.Writer, m *geometry.Mesh) error { return writeMeshSCADPolyhedron(w, m) }, nil
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

	mesh := geometry.GenerateMesh(&surface, cfg.meshOpts...)

	cfg.formatter(cfg.outFile, &mesh)

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

var meshCmd cli.Command = cli.Command{
	Name: "mesh",
	Usage: "Generate a mesh from elevation data\n" +
		"\n" +
		"Default GRID_REFERENCE is Snowdon summit (" + snowdon + ")",
	ArgsUsage: "GRID_REFERENCE",
	Flags: []cli.Flag{
		elevationFlag(),
		formatsFlag([]string{"scad"}),
		hscaleFlag(),
		vscaleFlag(),
		outfileFlag(true),
		widthFlag(),
	},
	Action: runMesh,
}
