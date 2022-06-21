package surface

import (
	"math"

	"github.com/usedbytes/osgrid"
	"github.com/usedbytes/osgrid/osdata"
)

type Surface struct {
	Data     [][]float64
	Max, Min float64
}

// Note that this will generate one more row/column than you might expect, as
// points form the corners of regions of the surface, not the centres
func Generate(db osdata.Float64Database, bottomLeft osgrid.GridRef,
	width, height osgrid.Distance, optsTODO ...interface{}) (Surface, error) {

	// TODO: Different resolutions
	resolution := db.Precision()

	nrows := height / resolution
	ncols := width / resolution
	data := make([][]float64, 0, nrows)

	maxElevation := float64(0.0)
	minElevation := math.MaxFloat64

	for north := osgrid.Distance(0); north <= height; north += resolution {
		row := make([]float64, 0, ncols)

		for east := osgrid.Distance(0); east <= width; east += resolution {
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

	return Surface{
		Data: data,
		Max:  maxElevation,
		Min:  minElevation,
	}, nil
}
