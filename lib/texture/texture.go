package texture

import (
	"image"
	"image/draw"
	"log"
	"math"

	"github.com/usedbytes/osgrid"
	"github.com/usedbytes/osgrid/osdata"
)

type Texture struct {
	Image image.Image
}

type GenerateTextureOpt func(t *Texture)

func GenerateTexture(db osdata.ImageDatabase, centre osgrid.GridRef,
	width, height osgrid.Distance, opts ...GenerateTextureOpt) (Texture, error) {

	bottomLeft, err := centre.Add(osgrid.Distance(-width/2), osgrid.Distance(-height/2))
	if err != nil {
		log.Fatal(err)
	}

	bottomRight, err := centre.Add(osgrid.Distance(width/2), osgrid.Distance(-height/2))
	if err != nil {
		log.Fatal(err)
	}

	topRight, err := centre.Add(osgrid.Distance(width/2), osgrid.Distance(height/2))
	if err != nil {
		log.Fatal(err)
	}

	tile, err := db.GetImageTile(bottomLeft)
	if err != nil {
		return Texture{}, err
	}

	// FIXME:
	//scale := 2.5
	scale := 1.0

	pixelWidth := int(math.Round(float64((width)) / scale))
	//downScale := 10
	//pixelWidth /= downScale
	canvas := image.NewRGBA(image.Rect(0, 0, pixelWidth, pixelWidth))

	drawnMinY := pixelWidth
	drawnMaxX := 0

	rowStart := bottomLeft

	for drawnMinY > 0 {
		dy := 0

		log.Println("top rowStart:", rowStart)
		coord := rowStart

		for drawnMaxX < pixelWidth {
			log.Println("drawnMaxX:", drawnMaxX, "pixelWidth:", pixelWidth)
			log.Println("drawnMinY:", drawnMinY)
			log.Println("coord:", coord)

			tile, err := db.GetImageTile(coord)
			if err != nil {
				log.Fatal("GetImageTile: ", err)
			}

			log.Println("tile:", tile.BottomLeft())

			img := tile.GetImage()

			// Pixel coordinate of the bottom left of this patch
			minX, maxY, err := tile.GetPixelCoord(coord)
			log.Println("minX, maxY:", minX, maxY)

			// How far right can we go within this tile?
			// Assume we need the whole width to start with
			maxX := img.Bounds().Dx()
			// Then check if the right hand edge is actually within this tile
			tileBottomRight, _ := tile.BottomLeft().Add(tile.Width(), 0)
			if tileBottomRight.Tile() == bottomRight.Tile() && tileBottomRight.TileEasting() > bottomRight.TileEasting() {
				// Find the pixel coordinate of the right edge.
				eastDistance := bottomRight.TileEasting() - tile.BottomLeft().TileEasting()
				regionRightEdge, _ := tile.BottomLeft().Add(eastDistance, 0)

				maxX, _, err = tile.GetPixelCoord(regionRightEdge)
				if err != nil {
					log.Fatalln("bottom right should be in tile:", err)
				}
			}
			log.Println("maxX:", maxX)

			// How far up can we go within this tile?
			// Assume we need the whole height to start with
			minY := 0
			// Then check if the top edge is actually within this tile
			tileTopRight, _ := tile.BottomLeft().Add(0, tile.Height())
			if tileTopRight.Tile() == topRight.Tile() && tileTopRight.TileNorthing() > topRight.TileNorthing() {
				// Find the pixel coordinate of the top edge.
				northDistance := topRight.TileNorthing() - tile.BottomLeft().TileNorthing()
				regionTopEdge, _ := tile.BottomLeft().Add(0, northDistance)

				_, minY, err = tile.GetPixelCoord(regionTopEdge)
				if err != nil {
					log.Fatalln("top right should be in tile:", err)
				}
			}
			log.Println("minY:", minY)

			//minX /= downScale
			//maxX /= downScale
			//minY /= downScale
			//maxY /= downScale
			//img := resize.Resize(img.Bounds().Dx() / downScale, img.Bounds.Dy() / downScale, img, resize.Lanczos3)

			// Copy Rect(minX, minY, maxX, maxY)
			dp := image.Pt(drawnMaxX, drawnMinY-(maxY-minY))
			sr := image.Rect(minX, minY, maxX, maxY)
			dr := image.Rectangle{dp, dp.Add(sr.Size())}
			draw.Draw(canvas, dr, img, sr.Min, draw.Src)

			log.Println("copy:", sr, "->", dr)

			drawnMaxX += sr.Size().X
			dy = sr.Size().Y

			coord, err = tile.BottomLeft().Add(tile.Width(), coord.TileNorthing()-tile.BottomLeft().TileNorthing())
			if err != nil {
				log.Fatal(err)
			}
		}

		// Start the next row
		drawnMaxX = 0
		drawnMinY -= dy

		//coord, err = tile.BottomLeft().Add(tile.Width(), coord.TileNorthing() - tile.BottomLeft().TileNorthing())
		rowStart, err = rowStart.Add(0, tile.Height())
		if err != nil {
			log.Fatal(err)
		}
		aligned := rowStart.Align(tile.Height())
		rowStart, _ = rowStart.Add(0, -(rowStart.TileNorthing() - aligned.TileNorthing()))
		//rowStart, _ = rowStart.Align(tile.Height()).Add(bottomLeft.TileEasting() - rowStart.TileEasting(), 0)
		log.Println("New row", drawnMinY)
		log.Println("New rowStart", rowStart)
	}

	return Texture{
		Image: canvas,
	}, nil
}
