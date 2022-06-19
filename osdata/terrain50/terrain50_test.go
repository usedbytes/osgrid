package terrain50

import (
	"strings"
	"testing"

	"github.com/usedbytes/osgrid"
)

var testASCData string = `
ncols 5
nrows 5
xllcorner 10000
yllcorner 20000
cellsize 2000
21.0 22.0 23.0 24.0 25.0
16.0 17.0 18.0 19.0 20.0
11.0 12.0 13.0 14.0 15.0
6.0 7.0 8.0 9.0 10.0
1.0 2.0 3.0 4.0 5.0`

func TestParseASCTile(t *testing.T) {
	r := strings.NewReader(testASCData)

	tile, err := ParseASCTile(r)
	if err != nil {
		t.Error(err)
	}

	sv12, _ := osgrid.ParseGridRef("SV 12")
	if tile.bottomLeft != sv12 {
		t.Errorf("bottomLeft: expected %s, got %s", sv12, tile.bottomLeft)
	}

	if tile.width != 10 * osgrid.Kilometre {
		t.Errorf("width: expected %d, got %d", 10 * osgrid.Kilometre, tile.width)
	}

	if tile.height != 10 * osgrid.Kilometre {
		t.Errorf("height: expected %d, got %d", 10 * osgrid.Kilometre, tile.height)
	}

	if tile.precision != 2 * osgrid.Kilometre {
		t.Errorf("precision: expected %d, got %d", 2 * osgrid.Kilometre, tile.precision)
	}

	for i, row := range tile.data {
		for j, v := range row {
			expect := float32(i * len(row) + j + 1)
			if v != expect {
				t.Errorf("(%d, %d): Expected %f, got %f\n", j, i, expect, v)
			}
		}
	}
}

func TestTileGet(t *testing.T) {
	r := strings.NewReader(testASCData)
	tile, err := ParseASCTile(r)
	if err != nil {
		t.Error(err)
	}

	sv1222, _ := osgrid.ParseGridRef("SV 1222")
	val, err := tile.Get(sv1222)
	if err != nil {
		t.Error(err)
	}

	if val != 7.0 {
		t.Errorf("Get: expected %f got %f", 7.0, val)
	}
}

/*
func TestOpenDatabase(t *testing.T) {
	_, err := OpenDatabase("/aux/data/os_terrain", 10 * osgrid.Kilometre)
	if err != nil {
		t.Error(err)
	}
}

func TestUseDatabase(t *testing.T) {
	d, err := OpenDatabase("/aux/data/os_terrain", 10 * osgrid.Kilometre)
	if err != nil {
		t.Error(err)
	}

	ref, _ := osgrid.ParseGridRef("SO 123578")
	val, err := d.GetData(ref)
	if err != nil {
		t.Error(err)
	}
	t.Log(ref, val)

	ref, _ = osgrid.ParseGridRef("SO 124579")
	val, err = d.GetData(ref)
	if err != nil {
		t.Error(err)
	}
	t.Log(ref, val)

	ref, _ = osgrid.ParseGridRef("SN 124579")
	val, err = d.GetData(ref)
	if err != nil {
		t.Error(err)
	}
	t.Log(ref, val)

	ref, _ = osgrid.ParseGridRef("NY 124579")
	val, err = d.GetData(ref)
	if err != nil {
		t.Error(err)
	}
	t.Log(ref, val)
}
*/
