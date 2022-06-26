package texture

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"testing"

	"github.com/usedbytes/osgrid"
	"github.com/usedbytes/osgrid/osdata"
)

type TestImageDatabase struct {
	resolution osgrid.Distance
	tileSize   osgrid.Distance
}

type TestImageTile struct {
	color      color.Color
	width      osgrid.Distance
	precision  osgrid.Distance
	bottomLeft osgrid.GridRef

	img image.Image
}

func (t *TestImageTile) String() string {
	return t.bottomLeft.String()
}

func (t *TestImageTile) GridRef() osgrid.GridRef {
	return t.bottomLeft
}

func (t *TestImageTile) BottomLeft() osgrid.GridRef {
	return t.bottomLeft
}

func (t *TestImageTile) Precision() osgrid.Distance {
	return t.precision
}

func (t *TestImageTile) Width() osgrid.Distance {
	return t.width
}

func (t *TestImageTile) Height() osgrid.Distance {
	return t.width
}

func (t *TestImageTile) GetImage() image.Image {
	return t.img
}

func (t *TestImageTile) GetPixelCoord(ref osgrid.GridRef) (int, int, error) {
	if ref.Align(t.width) != t.bottomLeft {
		return -1, -1, fmt.Errorf("Coordinate outside tile")
	}

	if t.precision != 1 {
		panic("TestImageTile only works for precision 1m")
	}

	ref = ref.Align(t.precision)

	east := ref.TileEasting() - t.bottomLeft.TileEasting()
	north := ref.TileNorthing() - t.bottomLeft.TileNorthing()

	// Works for 1 pixel per metre.
	return int(east), t.img.Bounds().Dy() - int(north), nil
}

func (db *TestImageDatabase) GetImageTile(ref osgrid.GridRef) (osdata.ImageTile, error) {
	ref = ref.Align(db.tileSize)

	absE, absN := ref.AbsEasting(), ref.AbsNorthing()
	eIdx := int(absE / db.tileSize)
	nIdx := int(absN / db.tileSize)

	g := uint8(((eIdx % 5) + 1) * 50)
	b := uint8((((nIdx + 3) % 5) + 1) * 50)

	img := image.NewNRGBA(image.Rect(0, 0, int(db.tileSize), int(db.tileSize)))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.RGBA{0, g, b, 0xff}), image.Pt(0, 0), draw.Over)

	return &TestImageTile{
		color:      color.RGBA{0, g, b, 0xff},
		width:      db.tileSize,
		precision:  1,
		bottomLeft: ref,
		img:        img,
	}, nil
}

func (db *TestImageDatabase) GetTile(ref osgrid.GridRef) (osdata.Tile, error) {
	return db.GetImageTile(ref)
}

func (db *TestImageDatabase) Precision() osgrid.Distance {
	return 1 * osgrid.Metre
}

func TestGenerateSingleTile(t *testing.T) {
	db := &TestImageDatabase{
		resolution: 1 * osgrid.Metre,
		tileSize:   10 * osgrid.Metre,
	}

	ref, err := osgrid.Origin().Add(5, 5)
	if err != nil {
		t.Fatal(err)
	}

	tex, err := GenerateTexture(db, ref, 10, 10)
	if err != nil {
		t.Fatal(err)
	}

	img := tex.Image

	if img.Bounds().Dx() != 10 || img.Bounds().Dy() != 10 {
		t.Errorf("bounds not correct, expected %v,%v got %v,%v", 10, 10, img.Bounds().Dx(), img.Bounds().Dy())
	}

	tile, err := db.GetImageTile(ref)
	if err != nil {
		t.Fatal(err)
	}
	ttile := tile.(*TestImageTile)
	er, eg, eb, _ := ttile.color.RGBA()

	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dy(); x++ {
			gr, gg, gb, _ := img.At(x, y).RGBA()

			if gr != er || gg != eg || gb != eb {
				t.Errorf("color at (%v,%v) not correct, expected (%v,%v,%v) got (%v,%v,%v)",
					x, y, er, eg, eb, gr, gg, gb)
			}
		}
	}

	if t.Failed() {
		f, err := os.Create("TestGenerateSingleTile.png")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		png.Encode(f, img)
	}
}

func TestGenerateMultiTile(t *testing.T) {
	db := &TestImageDatabase{
		resolution: 1 * osgrid.Metre,
		tileSize:   10 * osgrid.Metre,
	}

	ref, err := osgrid.Origin().Add(15, 15)
	if err != nil {
		t.Fatal(err)
	}

	tex, err := GenerateTexture(db, ref, 30, 30)
	if err != nil {
		t.Fatal(err)
	}

	img := tex.Image

	if img.Bounds().Dx() != 30 || img.Bounds().Dy() != 30 {
		t.Errorf("bounds not correct, expected %v,%v got %v,%v", 30, 30, img.Bounds().Dx(), img.Bounds().Dy())
	}

	for south := osgrid.Distance(0); south < 30; south += 10 {
		for east := osgrid.Distance(0); east < 30; east += 10 {
			ref, err := osgrid.Origin().Add(east, 20-south)
			if err != nil {
				t.Fatal(err)
			}

			tile, err := db.GetImageTile(ref)
			if err != nil {
				t.Fatal(err)
			}

			ttile := tile.(*TestImageTile)
			er, eg, eb, _ := ttile.color.RGBA()

			for y := 0; y < 10; y++ {
				for x := 0; x < 10; x++ {
					gr, gg, gb, _ := img.At(x+int(east), y+int(south)).RGBA()

					if gr != er || gg != eg || gb != eb {
						t.Errorf("color at (%v,%v) not correct, expected (%x,%x,%x) got (%x,%x,%x)",
							x+int(east), y+int(south), er, eg, eb, gr, gg, gb)
					}
				}
			}
		}
	}

	if t.Failed() {
		f, err := os.Create("TestGenerateMultiTile.png")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		png.Encode(f, img)
	}
}
