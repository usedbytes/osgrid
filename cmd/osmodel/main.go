package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

const snowdon = "SH 60986 54375"

// Re-use these between commands
// They aren't global options because then they aren't visible in the individual
// commands' help
func outfileFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:        "outfile",
		Aliases:     []string{"o"},
		Usage:       "`FILE` to write output to",
		Value:       "-",
		DefaultText: "stdout",
	}
}

func elevationFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:     "elevation",
		Aliases:  []string{"e"},
		Usage:    "`PATH` to elevation data (should contain 'data' folder)",
		Required: true,
		EnvVars:  []string{"OSMODEL_ELEVATION_DB"},
	}
}

func formatsFlag(formats []string) *cli.StringFlag {
	return &cli.StringFlag{
		Name:        "format",
		Aliases:     []string{"f"},
		Usage:       "`FORMAT` for output data. Allowed formats: " + strings.Join(formats, ", "),
		DefaultText: "from outfile extension",
	}
}

func widthFlag() *cli.UintFlag {
	return &cli.UintFlag{
		Name:        "width",
		Aliases:     []string{"w"},
		Usage:       "`WIDTH` and height in metres of the map area",
		Value:       5000,
		DefaultText: "5000 m",
	}
}

func main() {
	app := &cli.App{
		Name:  "osmodel",
		Usage: "Topographical model generator from Ordnance Survey open data",
		Commands: []*cli.Command{
			&surfaceCmd,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}
