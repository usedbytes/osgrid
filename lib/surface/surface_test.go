package surface

import (
	"errors"
	"strings"
	"testing"

	"github.com/usedbytes/osgrid"
	"github.com/usedbytes/osgrid/osdata"
)

type TestDatabase struct {
	resolution osgrid.Distance
}

func (db *TestDatabase) GetTile(osgrid.GridRef) (osdata.Tile, error) {
	return nil, errors.New("GetTile not implemented")
}

func (db *TestDatabase) Precision() osgrid.Distance {
	if db.resolution == 0 {
		return 1 * osgrid.Metre
	}

	return db.resolution
}

func (db *TestDatabase) GetFloat64(ref osgrid.GridRef) (float64, error) {
	return float64(ref.TileEasting() + ref.TileNorthing()), nil
}

func (db *TestDatabase) GetFloat64Tile(ref osgrid.GridRef) (osdata.Float64Tile, error) {
	return nil, errors.New("GetFloat64Tile not implemented")
}

func TestGenerateSimple(t *testing.T) {
	db := &TestDatabase{}

	s, err := Generate(db, osgrid.Origin(), 100*osgrid.Metre, 100*osgrid.Metre)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	expMin := float64(0.0)
	if s.Min != expMin {
		t.Errorf("Expected Min == %v, got %v", expMin, s.Min)
	}

	expMax := float64(100 + 100)
	if s.Max != expMax {
		t.Errorf("Expected Max == %v, got %v", expMax, s.Max)
	}

	if len(s.Data) != 101 {
		t.Errorf("Expected 101 rows, got %v", len(s.Data))
	}

	for y, row := range s.Data {
		if len(row) != 101 {
			t.Fatalf("Expected 101 cols, got %v", len(row))
		}
		for x, v := range row {
			if v != float64(x+y) {
				t.Fatalf("(%v, %v) expected %v got %v", x, y, float64(x+y), v)
			}
		}
	}
}

func TestGenerateResolution(t *testing.T) {
	db := &TestDatabase{}

	res := 20 * osgrid.Metre

	s, err := Generate(db, osgrid.Origin(), 100*osgrid.Metre, 100*osgrid.Metre, ResolutionOpt(res))
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if s.Resolution != res {
		t.Errorf("unexpected resolution, expected %v got %v", res, s.Resolution)
	}

	expRows := (100 / int(res)) + 1
	if len(s.Data) != expRows {
		t.Errorf("Expected %v rows, got %v", expRows, len(s.Data))
	}

	expCols := (100 / int(res)) + 1
	for y, row := range s.Data {
		if len(row) != expCols {
			t.Fatalf("Expected %v cols, got %v", expCols, len(row))
		}
		for x, v := range row {
			exp := float64(x*int(res) + y*int(res))
			if v != exp {
				t.Fatalf("(%v, %v) expected %v got %v", x, y, exp, v)
			}
		}
	}
}

func TestGenerateInvalidResolution(t *testing.T) {
	db := &TestDatabase{3 * osgrid.Metre}

	_, err := Generate(db, osgrid.Origin(), 100*osgrid.Metre, 100*osgrid.Metre, ResolutionOpt(20*osgrid.Metre))
	if err == nil {
		t.Fatal("Generate should have failed with invalid resolution")
	}

	if !strings.Contains(err.Error(), "multiple of") {
		t.Error("wrong error:", err)
	}

	_, err = Generate(db, osgrid.Origin(), 100*osgrid.Metre, 100*osgrid.Metre, ResolutionOpt(2*osgrid.Metre))
	if err == nil {
		t.Fatal("Generate should have failed with invalid resolution")
	}

	if !strings.Contains(err.Error(), "at least") {
		t.Error("wrong error:", err)
	}
}
