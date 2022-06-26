package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/usedbytes/osgrid"
	"github.com/usedbytes/osgrid/lib/texture"
	"github.com/usedbytes/osgrid/osdata"
	"github.com/usedbytes/osgrid/osdata/raster"
)

type TextureFormatter func(io.Writer, image.Image) error

func textureFormatterFromFormat(format string, c *cli.Context) (TextureFormatter, error) {
	switch format {
	case "png":
		return png.Encode, nil
	case "jpg":
		return func(w io.Writer, m image.Image) error { return jpeg.Encode(w, m, &jpeg.Options{Quality: 95}) }, nil
	default:
		return nil, fmt.Errorf("unknown format: %s", format)
	}
}

type textureConfig struct {
	rasterDB    osdata.ImageDatabase
	width       osgrid.Distance
	outFile     io.WriteCloser
	gridRef     osgrid.GridRef
	formatter   TextureFormatter
	textureOpts []texture.GenerateTextureOpt
}

func parseTextureArgs(c *cli.Context) (textureConfig, error) {
	var cfg textureConfig
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

	// raster
	cfg.rasterDB, err = raster.OpenDatabase(c.String("raster"), 10*osgrid.Kilometre)
	if err != nil {
		return textureConfig{}, fmt.Errorf("opening raster database: %w", err)
	}

	// GRID_REFERENCE
	gridRef := snowdon
	if c.NArg() > 0 {
		gridRef = strings.Join(c.Args().Slice(), "")
	}

	cfg.gridRef, err = osgrid.ParseGridRef(gridRef)
	if err != nil {
		return textureConfig{}, fmt.Errorf("parsing GRID_REFERENCE: %w", err)
	}

	// width
	cfg.width = osgrid.Distance(c.Uint("width"))

	// outfile
	cfg.outFile, err = os.Create(c.String("outfile"))
	if err != nil {
		return textureConfig{}, fmt.Errorf("opening outfile: %w", err)
	}
	cleanup = append(cleanup, func() { cfg.outFile.Close() })

	format := path.Ext(c.String("outfile"))[1:]

	// format (overrides file extension if set)
	if c.IsSet("format") {
		format = c.String("format")
	}

	cfg.formatter, err = textureFormatterFromFormat(format, c)
	if err != nil {
		return textureConfig{}, err
	}

	success = true

	return cfg, nil
}

func runTexture(c *cli.Context) error {
	cfg, err := parseTextureArgs(c)
	if err != nil {
		return err
	}
	defer cfg.outFile.Close()

	tex, err := texture.GenerateTexture(cfg.rasterDB, cfg.gridRef, cfg.width, cfg.width, cfg.textureOpts...)
	if err != nil {
		return fmt.Errorf("generating texture: %w", err)
	}

	cfg.formatter(cfg.outFile, tex.Image)

	return nil
}

var textureCmd cli.Command = cli.Command{
	Name: "texture",
	Usage: "Generate an image from raster data\n" +
		"\n" +
		"Default GRID_REFERENCE is Snowdon summit (" + snowdon + ")",
	ArgsUsage: "GRID_REFERENCE",
	Flags: []cli.Flag{
		rasterFlag(),
		formatsFlag([]string{"jpeg", "png"}),
		outfileFlag(false),
		widthFlag(),
	},
	Action: runTexture,
}
