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

func validateSimpleCube(t *testing.T, m *Mesh) {
	if len(m.Triangles) != (2 * 6) {
		t.Errorf("incorrect number of triangles. Expected %v, got %v", 2*6, len(m.Triangles))
	}

	if len(m.Vertices) != 8 {
		t.Errorf("incorrect number of vertices. Expected %v, got %v", 8, len(m.Vertices))
	}

	expZ := []float64{1, 2, 3, 4}
	for i, v := range m.Vertices[:4] {
		expX := float64(i%2) * m.HScale
		if v[0] != expX {
			t.Errorf("incorrect X on vertex %v. Expected %v, got %v", i, expX, v[0])
		}

		expY := float64(i/2) * m.HScale
		if v[1] != expY {
			t.Errorf("incorrect Y on vertex %v. Expected %v, got %v", i, expY, v[1])
		}

		if v[2] != expZ[i]*m.VScale {
			t.Errorf("incorrect Z on vertex %v. Expected %v, got %v", i, expZ[i]*m.VScale, v[2])
		}
	}

	for i, v := range m.Vertices[4:] {
		expX := float64(i%2) * m.HScale
		if v[0] != expX {
			t.Errorf("incorrect X on vertex %v. Expected %v, got %v", i, expX, v[0])
		}

		expY := float64(i/2) * m.HScale
		if v[1] != expY {
			t.Errorf("incorrect Y on vertex %v. Expected %v, got %v", i, expY, v[1])
		}

		expZ := float64(0)
		if v[2] != expZ {
			t.Errorf("incorrect Z on vertex %v. Expected %v, got %v", i, expZ, v[2])
		}
	}

	expTrisCW := [][3]uint{
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

	expTrisCCW := [][3]uint{
		// Top, ccw
		{0, 1, 3},
		{3, 2, 0},
		// Base, clockwise
		{4, 6, 7},
		{7, 5, 4},
		// South, ccw
		{4, 5, 1},
		{1, 0, 4},
		// North, ccw
		{7, 6, 2},
		{2, 3, 7},
		// East, ccw
		{5, 7, 3},
		{3, 1, 5},
		// West, ccw
		{6, 4, 0},
		{0, 2, 6},
	}

	expTris := expTrisCW
	if m.WindingCCW {
		expTris = expTrisCCW
	}

	for i, tri := range m.Triangles {
		if expTris[i][0] != tri[0] || expTris[i][1] != tri[1] || expTris[i][2] != tri[2] {
			t.Errorf("incorrect triangle %v. Expected %v, got %v", i, expTris[i], tri)
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

	validateSimpleCube(t, &m)
}

func TestGenerateMeshHScale(t *testing.T) {
	surf := Surface{
		Data: [][]float64{
			{1, 2},
			{3, 4},
		},
		Min:        1,
		Max:        4,
		Resolution: 1,
	}

	m := GenerateMesh(&surf, MeshHScaleOpt(2.5))

	if m.HScale != 2.5 {
		t.Fatalf("wrong HScale, expected %v, got %v", 2.5, m.HScale)
	}

	validateSimpleCube(t, &m)
}

func TestGenerateMeshVScale(t *testing.T) {
	surf := Surface{
		Data: [][]float64{
			{1, 2},
			{3, 4},
		},
		Min:        1,
		Max:        4,
		Resolution: 1,
	}

	m := GenerateMesh(&surf, MeshVScaleOpt(0.1))

	if m.VScale != 0.1 {
		t.Fatalf("wrong VScale, expected %v, got %v", 0.1, m.VScale)
	}

	validateSimpleCube(t, &m)
}

func TestGenerateMeshWindingCCW(t *testing.T) {
	surf := Surface{
		Data: [][]float64{
			{1, 2},
			{3, 4},
		},
		Min:        1,
		Max:        4,
		Resolution: 1,
	}

	m := GenerateMesh(&surf, MeshWindingOpt(true))

	if !m.WindingCCW {
		t.Fatalf("wrong winding, expected CCW")
	}

	validateSimpleCube(t, &m)
}

func TestGenerateMeshTexCoords(t *testing.T) {
	surf := Surface{
		Data: [][]float64{
			{1, 2, 3},
			{4, 5, 6},
			{7, 8, 9},
		},
		Min:        1,
		Max:        9,
		Resolution: 1,
	}

	tc := [][][2]float64{
		{{0.0, 0.0}, {0.5, 0.0}, {1.0, 0.0}},
		{{0.0, 0.5}, {0.5, 0.5}, {1.0, 0.5}},
		{{0.0, 1.0}, {0.5, 1.0}, {1.0, 1.0}},
	}

	m := GenerateMesh(&surf, MeshTextureCoordsOpt(tc))

	if len(m.TexCoords) != len(m.Vertices) {
		t.Errorf("wrong number of texture coordinates, expected %v got %v", len(m.Vertices), len(m.TexCoords))
	}

	for i, got := range m.TexCoords {
		tcrow := (i / 3) % 3
		tccol := i % 3
		exp := tc[tcrow][tccol]

		if exp[0] != got[0] || exp[1] != got[1] {
			t.Errorf("wrong coordinate, vertex %v, expected (%f,%f) got (%f,%f)", i, exp[0], exp[1], got[0], got[1])
		}
	}
}
