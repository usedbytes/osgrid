package geometry

import (
	"testing"
)

func TestMakeTriangles(t *testing.T) {
	// Layout for a simple 4-vert quad
	l := indexLayout{
		bStart: 0,
		tStart: 2,
		step:   1,
		stride: 2,
		cols:   2,
		rows:   2,
	}

	tris := makeTriangles(&l, true)

	exp := [][3]uint{
		{0, 1, 3},
		{3, 2, 0},
	}

	if len(tris) != len(exp) {
		t.Errorf("incorrect number of triangles. Expected %v, got %v", len(exp), len(tris))
	}

	for i, tri := range tris {
		if exp[i][0] != tri[0] || exp[i][1] != tri[1] || exp[i][2] != tri[2] {
			t.Errorf("incorrect triangle %v. Expected %v, got %v", i, exp[i], tri)
		}
	}
}

func TestMakeTrianglesCW(t *testing.T) {
	// Layout for a simple 4-vert quad
	l := indexLayout{
		bStart: 0,
		tStart: 2,
		step:   1,
		stride: 2,
		cols:   2,
		rows:   2,
	}

	tris := makeTriangles(&l, false)

	exp := [][3]uint{
		{0, 2, 3},
		{3, 1, 0},
	}

	if len(tris) != len(exp) {
		t.Errorf("incorrect number of triangles. Expected %v, got %v", len(exp), len(tris))
	}

	for i, tri := range tris {
		if exp[i][0] != tri[0] || exp[i][1] != tri[1] || exp[i][2] != tri[2] {
			t.Errorf("incorrect triangle %v. Expected %v, got %v", i, exp[i], tri)
		}
	}
}

func TestGenerateMeshSimple(t *testing.T) {
	surf := Surface{
		Data: [][]float64{
			{1, 2},
			{3, 4},
		},
		Min:        1,
		Max:        4,
		Resolution: 1,
	}

	m := GenerateMesh(&surf)

	if len(m.Triangles) != (2 * 6) {
		t.Errorf("incorrect number of triangles. Expected %v, got %v", 2*6, len(m.Triangles))
	}

	if len(m.Vertices) != 8 {
		t.Errorf("incorrect number of vertices. Expected %v, got %v", 8, len(m.Vertices))
	}

	expZ := []float64{1, 2, 3, 4}
	for i, v := range m.Vertices[:4] {
		expX := float64(i % 2)
		if v[0] != expX {
			t.Errorf("incorrect X on vertex %v. Expected %v, got %v", i, expX, v[0])
		}

		expY := float64(i / 2)
		if v[1] != expY {
			t.Errorf("incorrect Y on vertex %v. Expected %v, got %v", i, expY, v[1])
		}

		if v[2] != expZ[i] {
			t.Errorf("incorrect Z on vertex %v. Expected %v, got %v", i, expZ[i], v[2])
		}
	}

	for i, v := range m.Vertices[4:] {
		expX := float64(i % 2)
		if v[0] != expX {
			t.Errorf("incorrect X on vertex %v. Expected %v, got %v", i, expX, v[0])
		}

		expY := float64(i / 2)
		if v[1] != expY {
			t.Errorf("incorrect Y on vertex %v. Expected %v, got %v", i, expY, v[1])
		}

		expZ := float64(0)
		if v[2] != expZ {
			t.Errorf("incorrect Z on vertex %v. Expected %v, got %v", i, expZ, v[2])
		}
	}

	expTris := [][3]uint{
		// Top, clockwise
		{0, 2, 3},
		{3, 1, 0},
		// Base, ccw
		{4, 5, 7},
		{7, 6, 4},
		// South, clockwise
		{4, 0, 1},
		{1, 5, 4},
		// North, clockwise
		{7, 3, 2},
		{2, 6, 7},
		// East, clockwise
		{5, 1, 3},
		{3, 7, 5},
		// West, clockwise
		{6, 2, 0},
		{0, 4, 6},
	}

	for i, tri := range m.Triangles {
		if expTris[i][0] != tri[0] || expTris[i][1] != tri[1] || expTris[i][2] != tri[2] {
			t.Errorf("incorrect triangle %v. Expected %v, got %v", i, expTris[i], tri)
		}
	}
}
