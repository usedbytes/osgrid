package geometry

import (
	"fmt"
	"math"

	"github.com/usedbytes/osgrid"
	"github.com/usedbytes/osgrid/osdata"
)

type Surface struct {
	// Data[0][0] is the South-West corner
	Data       [][]float64
	Max, Min   float64
	Resolution osgrid.Distance
}

// Adjust the surface points so that 'max' becomes the new Max value
func (s *Surface) AdjustMax(max float64) {
	adjust := max - s.Max
	for _, row := range s.Data {
		for i, _ := range row {
			row[i] += adjust
		}
	}

	s.Min += adjust
	s.Max = max
}

// Multiply all the points in the surface by scale
func (s *Surface) Scale(scale float64) {
	for _, row := range s.Data {
		for i, _ := range row {
			row[i] *= scale
		}
	}

	s.Min *= scale
	s.Max *= scale
}

type GenerateSurfaceOpt func(*Surface)

func SurfaceResolutionOpt(res osgrid.Distance) GenerateSurfaceOpt {
	return func(s *Surface) {
		s.Resolution = res
	}
}

// Note that this will generate one more row/column than you might expect, as
// points form the corners of regions of the surface, not the centres
func GenerateSurface(db osdata.Float64Database, southWest osgrid.GridRef,
	width, height osgrid.Distance, opts ...GenerateSurfaceOpt) (Surface, error) {

	surf := Surface{
		Resolution: db.Precision(),
	}

	for _, opt := range opts {
		opt(&surf)
	}

	if surf.Resolution < db.Precision() {
		// TODO: This could be relaxed with some interpolation
		return Surface{}, fmt.Errorf("Resolution must be at least database precision (%v)", db.Precision())
	}

	if surf.Resolution%db.Precision() != 0 {
		// TODO: This could be relaxed with some interpolation
		return Surface{}, fmt.Errorf("Resolution must be a multiple of database precision (%v)", db.Precision())
	}

	nrows := height / surf.Resolution
	ncols := width / surf.Resolution
	data := make([][]float64, 0, nrows)

	maxElevation := float64(0.0)
	minElevation := math.MaxFloat64

	for north := osgrid.Distance(0); north <= height; north += surf.Resolution {
		row := make([]float64, 0, ncols)

		for east := osgrid.Distance(0); east <= width; east += surf.Resolution {
			ref, err := southWest.Add(east, north)
			if err != nil {
				return Surface{}, err
			}

			val, err := db.GetFloat64(ref)
			if err != nil {
				return Surface{}, err
			}

			if val > maxElevation {
				maxElevation = val
			}

			if val < minElevation {
				minElevation = val
			}

			row = append(row, val)
		}

		data = append(data, row)
	}

	surf.Data = data
	surf.Max = maxElevation
	surf.Min = minElevation

	return surf, nil
}
