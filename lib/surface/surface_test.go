package surface

import (
	"errors"
	"testing"

	"github.com/usedbytes/osgrid"
	"github.com/usedbytes/osgrid/osdata"
)

type TestDatabase struct {
}

func (db *TestDatabase) GetTile(osgrid.GridRef) (osdata.Tile, error) {
	return nil, errors.New("GetTile not implemented")
}

func (db *TestDatabase) Precision() osgrid.Distance {
	return 1 * osgrid.Metre
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
