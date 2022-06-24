package geometry

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

func TestGenerateSurfaceSimple(t *testing.T) {
	db := &TestDatabase{}

	s, err := GenerateSurface(db, osgrid.Origin(), 100*osgrid.Metre, 100*osgrid.Metre)
	if err != nil {
		t.Fatalf("GenerateSurface failed: %v", err)
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
			exp := float64(x + y)
			if v != exp {
				t.Fatalf("(%v, %v) expected %v got %v", x, y, exp, v)
			}
		}
	}
}

func TestGenerateSurfaceResolution(t *testing.T) {
	db := &TestDatabase{}

	res := 20 * osgrid.Metre

	s, err := GenerateSurface(db, osgrid.Origin(), 100*osgrid.Metre, 100*osgrid.Metre, SurfaceResolutionOpt(res))
	if err != nil {
		t.Fatalf("GenerateSurface failed: %v", err)
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

func TestGenerateSurfaceInvalidResolution(t *testing.T) {
	db := &TestDatabase{3 * osgrid.Metre}

	_, err := GenerateSurface(db, osgrid.Origin(), 100*osgrid.Metre, 100*osgrid.Metre, SurfaceResolutionOpt(20*osgrid.Metre))
	if err == nil {
		t.Fatal("GenerateSurface should have failed with invalid resolution")
	}

	if !strings.Contains(err.Error(), "multiple of") {
		t.Error("wrong error:", err)
	}

	_, err = GenerateSurface(db, osgrid.Origin(), 100*osgrid.Metre, 100*osgrid.Metre, SurfaceResolutionOpt(2*osgrid.Metre))
	if err == nil {
		t.Fatal("GenerateSurface should have failed with invalid resolution")
	}

	if !strings.Contains(err.Error(), "at least") {
		t.Error("wrong error:", err)
	}
}

func TestGenerateSurfaceNorthToSouth(t *testing.T) {
	db := &TestDatabase{}

	s, err := GenerateSurface(db, osgrid.Origin(), 100*osgrid.Metre, 100*osgrid.Metre, SurfaceNorthToSouthOpt(true))
	if err != nil {
		t.Fatalf("GenerateSurface failed: %v", err)
	}

	for y, _ := range s.Data {
		row := s.Data[len(s.Data)-y-1]
		if len(row) != 101 {
			t.Fatalf("Expected 101 cols, got %v", len(row))
		}
		for x, v := range row {
			exp := float64(x + y)
			if v != exp {
				t.Fatalf("(%v, %v) expected %v got %v", x, y, exp, v)
			}
		}
	}
}

func TestSurfaceAdjustMax(t *testing.T) {
	s := Surface{
		Data: [][]float64{
			{1.0, 2.0, 3.0, 4.0},
		},
		Max:        4.0,
		Min:        1.0,
		Resolution: 1 * osgrid.Metre,
	}

	newMax := 5.0
	s.AdjustMax(newMax)

	if s.Max != newMax {
		t.Errorf("expected max %v, got %v", newMax, s.Max)
	}

	if s.Min != (s.Max - 3.0) {
		t.Errorf("expected min %v, got %v", s.Max-3.0, s.Min)
	}

	exp := 2.0
	row := s.Data[0]
	for i, v := range row {
		if v != exp+float64(i) {
			t.Errorf("idx %v, expected %v, got %v", i, exp+float64(i), v)
		}
	}
}

func TestSurfaceScale(t *testing.T) {
	s := Surface{
		Data: [][]float64{
			{1.0, 2.0, 3.0, 4.0},
		},
		Max:        4.0,
		Min:        1.0,
		Resolution: 1 * osgrid.Metre,
	}

	scale := 1.25
	newMax := s.Max * scale
	newMin := s.Min * scale
	s.Scale(scale)

	if s.Max != newMax {
		t.Errorf("expected max %v, got %v", newMax, s.Max)
	}

	if s.Min != newMin {
		t.Errorf("expected min %v, got %v", newMin, s.Min)
	}

	orig := 1.0
	row := s.Data[0]
	for i, v := range row {
		if v != (orig+float64(i))*scale {
			t.Errorf("idx %v, expected %v, got %v", i, (orig+float64(i))*scale, v)
		}
	}
}
