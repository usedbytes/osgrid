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
	precision      osgrid.Distance
	pixelPrecision int
	tileSize       osgrid.Distance
}

type TestImageTile struct {
	color          color.Color
	width          osgrid.Distance
	precision      osgrid.Distance
	pixelPrecision int
	bottomLeft     osgrid.GridRef

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

func (t *TestImageTile) PixelPrecision() int {
	return t.pixelPrecision
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

	ref = ref.Align(t.precision)

	east := ref.TileEasting() - t.bottomLeft.TileEasting()
	north := ref.TileNorthing() - t.bottomLeft.TileNorthing()

	blocksEast := int(east / t.precision)
	blocksNorth := int(north / t.precision)

	pixelsEast := blocksEast * t.pixelPrecision
	pixelsNorth := blocksNorth * t.pixelPrecision

	return pixelsEast, t.img.Bounds().Dy() - pixelsNorth, nil
}

func (db *TestImageDatabase) GetImageTile(ref osgrid.GridRef) (osdata.ImageTile, error) {
	ref = ref.Align(db.tileSize)

	absE, absN := ref.AbsEasting(), ref.AbsNorthing()
	eIdx := int(absE / db.tileSize)
	nIdx := int(absN / db.tileSize)

	g := uint8(((eIdx % 5) + 1) * 50)
	b := uint8((((nIdx + 3) % 5) + 1) * 50)

	pixelSize := int(db.tileSize/db.precision) * db.pixelPrecision

	img := image.NewNRGBA(image.Rect(0, 0, pixelSize, pixelSize))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.RGBA{0, g, b, 0xff}), image.Pt(0, 0), draw.Over)

	return &TestImageTile{
		color:          color.RGBA{0, g, b, 0xff},
		width:          db.tileSize,
		precision:      db.precision,
		pixelPrecision: db.pixelPrecision,
		bottomLeft:     ref,
		img:            img,
	}, nil
}

func (db *TestImageDatabase) GetTile(ref osgrid.GridRef) (osdata.Tile, error) {
	return db.GetImageTile(ref)
}

func (db *TestImageDatabase) Precision() osgrid.Distance {
	return 1 * osgrid.Metre
}

func testSingleTile(t *testing.T, db *TestImageDatabase) {
	sizeMetres := db.tileSize

	ref, err := osgrid.Origin().Add(sizeMetres/2, sizeMetres/2)
	if err != nil {
		t.Fatal(err)
	}

	tile, err := db.GetImageTile(ref)
	if err != nil {
		t.Fatal(err)
	}

	sizePixels := osdata.DistanceToPixels(tile, sizeMetres)

	tex, err := GenerateTexture(db, ref, sizeMetres, sizeMetres)
	if err != nil {
		t.Fatal(err)
	}

	img := tex.Image

	if img.Bounds().Dx() != sizePixels || img.Bounds().Dy() != sizePixels {
		t.Errorf("bounds not correct, expected %v,%v got %v,%v", sizePixels, sizePixels, img.Bounds().Dx(), img.Bounds().Dy())
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

	if t.Failed() || testing.Verbose() {
		f, err := os.Create(t.Name() + ".png")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		png.Encode(f, img)
	}
}

func TestGenerateSingleTile(t *testing.T) {
	db := &TestImageDatabase{
		precision:      1 * osgrid.Metre,
		pixelPrecision: 1,
		tileSize:       10 * osgrid.Metre,
	}

	testSingleTile(t, db)
}

func testMultiTile(t *testing.T, db *TestImageDatabase) {
	sizeMetres := db.tileSize * 3

	ref, err := osgrid.Origin().Add(sizeMetres/2, sizeMetres/2)
	if err != nil {
		t.Fatal(err)
	}

	tile, err := db.GetImageTile(ref)
	if err != nil {
		t.Fatal(err)
	}

	sizePixels := osdata.DistanceToPixels(tile, sizeMetres)

	tex, err := GenerateTexture(db, ref, sizeMetres, sizeMetres)
	if err != nil {
		t.Fatal(err)
	}

	img := tex.Image

	if img.Bounds().Dx() != sizePixels || img.Bounds().Dy() != sizePixels {
		t.Errorf("bounds not correct, expected %v,%v got %v,%v", sizePixels, sizePixels, img.Bounds().Dx(), img.Bounds().Dy())
	}

	for south := osgrid.Distance(0); south < sizeMetres; south += db.tileSize {
		for east := osgrid.Distance(0); east < sizeMetres; east += db.tileSize {
			ref, err := osgrid.Origin().Add(east, sizeMetres-south-db.tileSize)
			if err != nil {
				t.Fatal(err)
			}

			tile, err := db.GetImageTile(ref)
			if err != nil {
				t.Fatal(err)
			}

			ttile := tile.(*TestImageTile)
			er, eg, eb, _ := ttile.color.RGBA()

			for y := 0; y < osdata.DistanceToPixels(tile, db.tileSize); y++ {
				for x := 0; x < osdata.DistanceToPixels(tile, db.tileSize); x++ {
					tx, ty := osdata.DistanceToPixels(tile, east)+x, osdata.DistanceToPixels(tile, south)+y
					gr, gg, gb, _ := img.At(tx, ty).RGBA()

					if gr != er || gg != eg || gb != eb {
						t.Errorf("color at (%v,%v) not correct, expected (%x,%x,%x) got (%x,%x,%x)",
							tx, ty, er, eg, eb, gr, gg, gb)
					}
				}
			}
		}
	}

	if t.Failed() || testing.Verbose() {
		f, err := os.Create(t.Name() + ".png")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		png.Encode(f, img)
	}
}

func TestGenerateMultiTile(t *testing.T) {
	db := &TestImageDatabase{
		precision:      1 * osgrid.Metre,
		pixelPrecision: 1,
		tileSize:       10 * osgrid.Metre,
	}

	testMultiTile(t, db)
}

func TestGenerateSingleTilePrecision(t *testing.T) {
	db := &TestImageDatabase{
		precision:      5 * osgrid.Metre,
		pixelPrecision: 2,
		tileSize:       10 * osgrid.Metre,
	}

	testSingleTile(t, db)
}

func TestGenerateMultiTilePrecision(t *testing.T) {
	db := &TestImageDatabase{
		precision:      5 * osgrid.Metre,
		pixelPrecision: 2,
		tileSize:       10 * osgrid.Metre,
	}

	testMultiTile(t, db)
}
