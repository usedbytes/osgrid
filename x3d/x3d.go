package main

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// AttrMarshaler
type MFVec3f [][3]float64

func (v MFVec3f) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	attr := xml.Attr{
		Name: name,
	}

	builder := &strings.Builder{}

	for i, v3 := range v {
		builder.WriteString(fmt.Sprintf("%f %f %f", v3[0], v3[1], v3[2]))
		if i < len(v) - 1 {
			builder.WriteRune(' ')
		}
	}

	attr.Value = builder.String()

	return attr, nil
}

type MFInt32 []int32

func (v MFInt32) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	attr := xml.Attr{
		Name: name,
	}

	builder := &strings.Builder{}

	for i, val := range v {
		builder.WriteString(fmt.Sprintf("%d", val))
		if i < len(v) - 1 {
			builder.WriteRune(' ')
		}
	}

	attr.Value = builder.String()

	return attr, nil
}

type X3DVersion float64
type ProfileNames string

type X3D struct {
	Version X3DVersion   `xml:"version,attr"`
	Profile ProfileNames `xml:"profile,attr"`
	Scene *Scene
}

type Scene struct {
	Shape *Shape
}

type Shape struct {
	IndexedFaceSet *IndexedFaceSet
}

type IndexedFaceSet struct {
	CCW           bool    `xml:"ccw,attr"`
	Convex        bool    `xml:"convex,attr"`
	CreaseAngle   float64 `xml:"creaseAngle,attr"`
	ColorIndex    MFInt32 `xml:"colorIndex,attr"`
	CoordIndex    MFInt32 `xml:"coordIndex,attr"`
	NormalIndex   MFInt32 `xml:"normalIndex,attr"`
	TexCoordIndex MFInt32 `xml:"texCoordIndex,attr"`
	Coordinate *Coordinate
}

type Coordinate struct {
	Point MFVec3f `xml:"point,attr"`
}

type Face MFInt32

type Mesh struct {
	// Points needs to be ordered bottom-left to top-right
	Points MFVec3f
	Width, Height int
}

func (m *Mesh) Triangles() MFInt32 {
	// It's (w * 2) - 2 triangles per row, and there's h - 1 rows.
	// We have to use 4 indices per triangle.
	faces := make(MFInt32, 0, ((2*m.Width-2)*(m.Height-1)))

	for y := 0; y < m.Height - 1; y++ {
		for x := 0; x < m.Width - 1; x++ {
			idx := y * m.Width + x

			v0, v1, v2, v3 := int32(idx), int32(idx + 1), int32(idx + m.Width), int32(idx + m.Width + 1)

			faces = append(faces, MFInt32{v0, v1, v3, -1}...)
			faces = append(faces, MFInt32{v3, v2, v0, -1}...)
		}
	}

	return faces
}
