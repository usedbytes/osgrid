package x3d

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
		if i < len(v)-1 {
			builder.WriteRune(' ')
		}
	}

	attr.Value = builder.String()

	return attr, nil
}

type MFVec2f [][2]float64

func (v MFVec2f) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	attr := xml.Attr{
		Name: name,
	}

	builder := &strings.Builder{}

	for i, v2 := range v {
		builder.WriteString(fmt.Sprintf("%f %f", v2[0], v2[1]))
		if i < len(v)-1 {
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
		if i < len(v)-1 {
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
	Scene   *Scene
}

type Scene struct {
	Shape *Shape
}

type TextureProperties struct {
	BoundaryModeS string `xml:"boundaryModeS,attr"`
	BoundaryModeT string `xml:"boundaryModeT,attr"`
}

type ImageTexture struct {
	URL               string `xml:"url,attr"`
	TextureProperties *TextureProperties
}

type Appearance struct {
	ImageTexture *ImageTexture
}

type Shape struct {
	IndexedFaceSet *IndexedFaceSet
	Appearance     *Appearance
}

type IndexedFaceSet struct {
	CCW               bool    `xml:"ccw,attr"`
	Convex            bool    `xml:"convex,attr"`
	CreaseAngle       float64 `xml:"creaseAngle,attr"`
	ColorIndex        MFInt32 `xml:"colorIndex,attr"`
	CoordIndex        MFInt32 `xml:"coordIndex,attr"`
	NormalIndex       MFInt32 `xml:"normalIndex,attr"`
	TexCoordIndex     MFInt32 `xml:"texCoordIndex,attr"`
	Coordinate        *Coordinate
	TextureCoordinate *TextureCoordinate
}

type Coordinate struct {
	Point MFVec3f `xml:"point,attr"`
}

type TextureCoordinate struct {
	Point MFVec2f `xml:"point,attr"`
}
