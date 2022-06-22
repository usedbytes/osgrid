package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/usedbytes/osgrid"
	"github.com/usedbytes/osgrid/lib/surface"
	"github.com/usedbytes/osgrid/osdata"
	"github.com/usedbytes/osgrid/osdata/terrain50"
)

type SurfaceFormatter func(io.Writer, *surface.Surface) error

type surfaceConfig struct {
	elevationDB osdata.Float64Database
	width       osgrid.Distance
	outFile     io.WriteCloser
	gridRef     osgrid.GridRef
	formatter   SurfaceFormatter
}

func surfaceFormatterFromFormat(format string, c *cli.Context) (SurfaceFormatter, error) {
	switch format {
	case "txt":
		return func(w io.Writer, s *surface.Surface) error { return writeSurfaceSV(w, s, c.String("sep")) }, nil
	case "csv":
		return func(w io.Writer, s *surface.Surface) error { return writeSurfaceSV(w, s, ",") }, nil
	case "tsv":
		return func(w io.Writer, s *surface.Surface) error { return writeSurfaceSV(w, s, "\t") }, nil
	case "dat":
		return func(w io.Writer, s *surface.Surface) error { return writeSurfaceSV(w, s, " ") }, nil
	default:
		return nil, fmt.Errorf("unknown format: %s", format)
	}
}

func parseSurfaceArgs(c *cli.Context) (surfaceConfig, error) {
	var cfg surfaceConfig
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
		return surfaceConfig{}, fmt.Errorf("opening elevation database: %w", err)
	}

	// GRID_REFERENCE
	gridRef := snowdon
	if c.NArg() > 0 {
		gridRef = strings.Join(c.Args().Slice(), "")
	}

	cfg.gridRef, err = osgrid.ParseGridRef(gridRef)
	if err != nil {
		return surfaceConfig{}, fmt.Errorf("parsing GRID_REFERENCE: %w", err)
	}

	// width
	cfg.width = osgrid.Distance(c.Uint("width"))

	format := "txt"

	// outfile
	if c.String("outfile") == "-" {
		cfg.outFile = os.Stdout
	} else {
		cfg.outFile, err = os.Create(c.String("outfile"))
		if err != nil {
			return surfaceConfig{}, fmt.Errorf("opening outfile: %w", err)
		}
		cleanup = append(cleanup, func() { cfg.outFile.Close() })

		format = path.Ext(c.String("outfile"))[1:]
	}

	// format (overrides file extension if set)
	if c.IsSet("format") {
		format = c.String("format")
	}

	cfg.formatter, err = surfaceFormatterFromFormat(format, c)
	if err != nil {
		return surfaceConfig{}, err
	}

	success = true

	return cfg, nil
}

func writeSurfaceSV(w io.Writer, s *surface.Surface, sep string) error {
	bld := &strings.Builder{}
	for _, row := range s.Data {
		for i, v := range row {
			bld.WriteString(fmt.Sprintf("%v", v))
			if i < len(row)-1 {
				bld.WriteString(sep)
			}
		}
		bld.WriteString("\n")

		n, err := io.WriteString(w, bld.String())
		if err != nil {
			return err
		} else if n != bld.Len() {
			return fmt.Errorf("short write")
		}
		bld.Reset()
	}

	return nil
}

func runSurface(c *cli.Context) error {
	cfg, err := parseSurfaceArgs(c)
	if err != nil {
		return err
	}
	defer cfg.outFile.Close()

	bottomLeft, err := cfg.gridRef.Add(-cfg.width/2, -cfg.width/2)
	if err != nil {
		return fmt.Errorf("map bounds: %w", err)
	}

	surface, err := surface.Generate(cfg.elevationDB, bottomLeft, cfg.width, cfg.width)
	if err != nil {
		return fmt.Errorf("generating surface: %w", err)
	}

	cfg.formatter(cfg.outFile, &surface)

	return nil
}

func sepFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:  "sep",
		Usage: "`SEPARATOR` for txt output",
		Value: ", ",
	}
}

var surfaceCmd cli.Command = cli.Command{
	Name: "surface",
	Usage: "Generate a surface from elevation data\n" +
		"\n" +
		"Default GRID_REFERENCE is Snowdon summit (" + snowdon + ")",
	ArgsUsage: "GRID_REFERENCE",
	Flags: []cli.Flag{
		elevationFlag(),
		formatsFlag([]string{"csv", "dat", "tsv", "txt"}),
		outfileFlag(),
		sepFlag(),
		widthFlag(),
	},
	Action: runSurface,
}