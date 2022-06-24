package geometry

import (
	"math"
)

type Mesh struct {
	Vertices   [][3]float64
	Triangles  [][3]uint
	WindingCCW bool
	HScale     float64
	VScale     float64
}

// Description of a mesh in terms of indices, to create triangles
type indexLayout struct {
	// The starting indices for the bottom/top of the first strip
	bStart, tStart int
	// Step between adjacent vertices in a strip
	step int
	// Stride between strips
	stride int

	// Number of columns and rows of vertices
	cols, rows int
}

func makeTriangles(l *indexLayout, ccw bool) [][3]uint {
	strips := l.rows - 1

	// There's (cols * 2) - 2 triangles per strip
	tris := make([][3]uint, 0, ((2*l.cols - 2) * strips))

	for strip := 0; strip < int(strips); strip++ {
		bIdx := l.bStart + strip*l.stride
		tIdx := l.tStart + strip*l.stride

		for col := 0; col < l.cols-1; col++ {
			v0, v1, v2, v3 := uint(bIdx), uint(bIdx+l.step), uint(tIdx), uint(tIdx+l.step)

			if ccw {
				tris = append(tris, [3]uint{v0, v1, v3})
				tris = append(tris, [3]uint{v3, v2, v0})
			} else {
				tris = append(tris, [3]uint{v0, v2, v3})
				tris = append(tris, [3]uint{v3, v1, v0})
			}

			bIdx += l.step
			tIdx += l.step
		}
	}

	return tris
}

// Options:
// HScale
// VScale --> No? Should apply to surface first.
// WallThickness
// Winding
//

type GenerateMeshOpt func(*Mesh)

func GenerateMesh(s *Surface, opts ...GenerateMeshOpt) Mesh {
	m := Mesh{
		HScale: 1.0,
		VScale: 1.0,
	}

	for _, opt := range opts {
		opt(&m)
	}

	rows := len(s.Data)
	cols := len(s.Data[0])
	hstep := float64(s.Resolution) * m.HScale

	// Top and bottom
	m.Vertices = make([][3]float64, rows*cols*2)

	for r, row := range s.Data {
		y := float64(r) * hstep

		for c, v := range row {
			x := float64(c) * hstep

			topIdx := r*cols + c
			baseIdx := cols*rows + topIdx

			m.Vertices[topIdx] = [3]float64{x, y, v * m.VScale}
			m.Vertices[baseIdx] = [3]float64{x, y, math.Min(s.Min, 0)}
		}
	}

	layout := indexLayout{
		bStart: 0,
		tStart: cols,
		step:   1,
		stride: cols,

		rows: rows,
		cols: cols,
	}
	topTris := makeTriangles(&layout, m.WindingCCW)
	m.Triangles = topTris

	// Base triangles
	// Winding is reversed so that the base faces outwards
	layout.bStart = rows * cols
	layout.tStart = rows*cols + layout.stride
	baseTris := makeTriangles(&layout, !m.WindingCCW)
	m.Triangles = append(m.Triangles, baseTris...)

	// South face
	layout.bStart = rows * cols
	layout.tStart = 0
	layout.rows = 2
	southTris := makeTriangles(&layout, m.WindingCCW)
	m.Triangles = append(m.Triangles, southTris...)

	// North face
	layout.bStart = (rows * cols * 2) - 1
	layout.tStart = (rows * cols) - 1
	layout.step = -1
	layout.rows = 2
	northTris := makeTriangles(&layout, m.WindingCCW)
	m.Triangles = append(m.Triangles, northTris...)

	// East face
	layout.bStart = (rows * cols) + cols - 1
	layout.tStart = cols - 1
	layout.step = cols
	layout.rows = 2
	eastTris := makeTriangles(&layout, m.WindingCCW)
	m.Triangles = append(m.Triangles, eastTris...)

	// West face
	layout.bStart = (rows * cols * 2) - cols
	layout.tStart = (rows * cols) - cols
	layout.step = -cols
	layout.rows = 2
	westTris := makeTriangles(&layout, m.WindingCCW)
	m.Triangles = append(m.Triangles, westTris...)

	return m
}
