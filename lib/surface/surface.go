package surface

import (
	"fmt"
	"math"

	"github.com/usedbytes/osgrid"
	"github.com/usedbytes/osgrid/osdata"
)

type Surface struct {
	Data       [][]float64
	Max, Min   float64
	Resolution osgrid.Distance
}

type GenerateOpt func(*Surface)

func ResolutionOpt(res osgrid.Distance) GenerateOpt {
	return func(s *Surface) {
		s.Resolution = res
	}
}

// Note that this will generate one more row/column than you might expect, as
// points form the corners of regions of the surface, not the centres
func Generate(db osdata.Float64Database, bottomLeft osgrid.GridRef,
	width, height osgrid.Distance, opts ...GenerateOpt) (Surface, error) {

	surf := Surface{
		Resolution: db.Precision(),
	}

	for _, opt := range opts {
		opt(&surf)
	}

	if surf.Resolution < db.Precision() {
		// TODO: This could be relaxed with some extrapolation
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
			ref, err := bottomLeft.Add(east, north)
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
