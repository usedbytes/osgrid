package terrain50

import (
	"archive/zip"
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"os"
	"regexp"
	"strings"
	"strconv"
	"github.com/usedbytes/osgrid"
)

type Tile struct {
	bottomLeft osgrid.GridRef
	width, height osgrid.Distance
	precision osgrid.Distance
	data [][]float32
}

type Database struct {
	path string
	tileSize osgrid.Distance
}

func parseTileData(r io.Reader) ([][]float32, error) {
	c := csv.NewReader(r)
	c.Comma = ' '

	records, err := c.ReadAll()
	if err != nil {
		fmt.Println("csverr")
		return nil, err
	}

	vals := make([][]float32, len(records))
	for i, _ := range vals {
		vals[i] = make([]float32, len(records[0]))
	}

	nrows := len(records)
	for y, row := range records {
		for x, v := range row {
			floatVal, err := strconv.ParseFloat(v, 32)
			if err != nil {
				return nil, err
			}
			// We need to flip the Y axis
			// The Terrain 50 data is NW to SE
			// But we want the origin in the bottom-left (SW to NE)
			vals[nrows - y - 1][x] = float32(floatVal)
		}
	}

	return vals, nil
}

func ParseASCTile(r io.Reader) (*Tile, error) {
	buf := bufio.NewReader(r)

	var err error
	var ncols, nrows int
	var xllcorner, yllcorner, cellsize osgrid.Distance

	header := true
	for ; header ; {
		line, err := buf.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)

		if len(line) == 0 {
			continue
		}

		fields := strings.Split(line, " ")
		if len(fields) != 2 {
			return nil, fmt.Errorf("Unexpected header data: %s", line)
		}
		value, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, err
		}

		switch fields[0] {
		case "ncols":
			ncols = value
		case "nrows":
			nrows = value
		case "xllcorner":
			xllcorner = osgrid.Distance(value)
		case "yllcorner":
			yllcorner = osgrid.Distance(value)
		case "cellsize":
			cellsize = osgrid.Distance(value)
			header = false
		}
	}

	var t Tile
	t.bottomLeft, err = osgrid.Origin().Add(xllcorner, yllcorner)
	if err != nil {
		return nil, err
	}
	t.width = osgrid.Distance(ncols) * cellsize * osgrid.Metre
	t.height = osgrid.Distance(nrows) * cellsize * osgrid.Metre
	t.precision = cellsize * osgrid.Metre

	if t.width == 0 || t.width != t.height {
		return nil, fmt.Errorf("Invalid tile size")
	}

	t.data, err = parseTileData(buf)
	if err != nil {
		return nil, err
	}

	if len(t.data) != nrows || len(t.data[0]) != ncols {
		return nil, fmt.Errorf("Invalid amount of data")
	}

	return &t, nil
}

func OpenTile(path string) (*Tile, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		// TODO: Not implemented
		return nil, fmt.Errorf("Tiles are expected to be zipped")
	}

	zipFile, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer zipFile.Close()

	t := &Tile{
	}

	for _, f := range zipFile.File {
		if filepath.Ext(f.Name) == ".asc" {
			r, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer r.Close()

			t, err = ParseASCTile(r)
			if err != nil {
				return nil, err
			}
			return t, nil
		}
	}

	return nil, nil
}

func (d *Database) findTile(ref osgrid.GridRef) (string, error) {
	ref.Align(d.tileSize)

	dirExp := fmt.Sprintf("(?i)%s", ref.Tile())
	dirRe := regexp.MustCompile(dirExp)

	dir, err := ioutil.ReadDir(d.path)
	if err != nil {
		return "", err
	}

	for _, dirEntry := range dir {
		if !dirEntry.IsDir() {
			continue
		}

		if dirRe.MatchString(dirEntry.Name()) {
			zipExp := fmt.Sprintf("(?i)%s%s", ref.Tile(), ref.Digits())
			zipRe := regexp.MustCompile(zipExp)

			zipDir, err := ioutil.ReadDir(filepath.Join(d.path, dirEntry.Name()))
			if err != nil {
				return "", err
			}

			for _, zipEntry := range zipDir {
				if zipRe.MatchString(zipEntry.Name()) {
					ret := filepath.Join(d.path, dirEntry.Name(), zipEntry.Name())
					return ret, nil
				}
			}
		}
	}

	return "", fmt.Errorf("Tile %s not found", ref)
}

func OpenDatabase(path string, tileSize osgrid.Distance) (*Database, error) {
	datapath := filepath.Join(path, "data")

	fi, err := os.Stat(datapath)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("%s should be a directory", datapath)
	}

	d := &Database{
		path: datapath,
		tileSize: tileSize,
	}

	// We assume that London is available in the data-set
	// TODO: That might be a bad assumption, but good enough for now
	tq28, _ := osgrid.ParseGridRef("TQ 28")
	loc, err := d.findTile(tq28)
	if err != nil {
		return nil, err
	}

	t, err := OpenTile(loc)
	if err != nil {
		return nil, err
	}

	if t.width != tileSize || t.height != tileSize {
		return nil, fmt.Errorf("Specified tileSize (%d) doesn't match data (%d)", tileSize, t.width)
	}

	return d, nil
}
