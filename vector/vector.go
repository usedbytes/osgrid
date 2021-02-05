package vector

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"image"
	"math"
	"os"
	"path/filepath"
	"regexp"

	"github.com/usedbytes/osgrid"
	"github.com/google/tiff"
	_ "golang.org/x/image/tiff"
)

type Tile struct {
	bottomLeft osgrid.GridRef
	width, height osgrid.Distance
	precision osgrid.Distance

	image image.Image
	scaleX, scaleY float64
}

func (t *Tile) String() string {
	return t.bottomLeft.String()
}

func (t *Tile) GridRef() osgrid.GridRef {
	return t.bottomLeft
}

func (t *Tile) Precision() osgrid.Distance {
	return t.precision
}

func (t *Tile) Width() osgrid.Distance {
	return t.width
}

func (t *Tile) Height() osgrid.Distance {
	return t.height
}

func (t *Tile) Image() image.Image {
	return t.image
}

func (t *Tile) GetPixelCoord(ref osgrid.GridRef) (int, int, error) {
	if ref.Align(t.width) != t.bottomLeft {
		return -1, -1, fmt.Errorf("Coordinate outside tile")
	}

	ref = ref.Align(t.precision)

	east := ref.TileEasting() - t.bottomLeft.TileEasting()
	north := ref.TileNorthing() - t.bottomLeft.TileNorthing()

	x := int(float64(east) / t.scaleX)
	y := t.image.Bounds().Dy() - int(float64(north) / t.scaleY)

	return x, y, nil
}

type tileMapEntry struct {
	path string
	slot int
}

type tileCacheEntry struct {
	timestamp int
	tile *Tile
}

type Database struct {
	path string
	tileSize osgrid.Distance
	precision osgrid.Distance

	tileMap map[string]tileMapEntry
	tileCache []tileCacheEntry
	timestamp int
}

const (
	ModelPixelScaleTag uint16 = 33550
	ModelTiepointTag   uint16 = 33922
)

type TiePoint struct {
	PixX, PixY int
	ModelX, ModelY int
}

type TiePointTag struct {
	TiePoints []TiePoint
}

func (tt *TiePointTag) Decode(data []byte, order binary.ByteOrder) error {
	numTies := len(data) / (8 * 6)

	// Should be a multiple of 6 doubles
	if numTies * (8 * 6) != len(data) {
		return fmt.Errorf("expected tie point to be multiple of 6 doubles")
	}

	tps := make([]TiePoint, numTies)

	for i := range tps {
		u64 := order.Uint64(data[(i * 8 * 6) + (8 * 0):])
		tps[i].PixX = int(math.Float64frombits(u64))
		u64 = order.Uint64(data[(i * 8 * 6) + (8 * 1):])
		tps[i].PixY = int(math.Float64frombits(u64))
		// Skip K

		u64 = order.Uint64(data[(i * 8 * 6) + (8 * 3):])
		tps[i].ModelX = int(math.Float64frombits(u64))
		u64 = order.Uint64(data[(i * 8 * 6) + (8 * 4):])
		tps[i].ModelY = int(math.Float64frombits(u64))
		// Skip Z
	}

	tt.TiePoints = tps

	return nil
}

type PixelScaleTag struct {
	ScaleX, ScaleY float64
}

func (pt *PixelScaleTag) Decode(data []byte, order binary.ByteOrder) error {
	if len(data) != 24 {
		return fmt.Errorf("expected scale to 3 doubles")
	}

	u64 := order.Uint64(data[(8 * 0):])
	pt.ScaleX = math.Float64frombits(u64)
	u64 = order.Uint64(data[(8 * 1):])
	pt.ScaleY = math.Float64frombits(u64)
	// Skip Z

	return nil
}

func OpenTile(path string) (*Tile, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return nil, fmt.Errorf("tile expected to be an image")
	}

	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}

	r.Seek(0, 0)

	// Parse out the tags
	t, err := tiff.Parse(r, nil, nil)
	if err != nil {
		return nil, err
	}

	ifds := t.IFDs()
	if len(ifds) != 1 {
		// Being lazy
		return nil, fmt.Errorf("can't handle more than one IFD")
	}

	if !ifds[0].HasField(ModelPixelScaleTag) || !ifds[0].HasField(ModelTiepointTag) {
		return nil, fmt.Errorf("need pixel scale and tie points")
	}

	scaleField := ifds[0].GetField(ModelPixelScaleTag)
	scaleVal := scaleField.Value()
	scaleTag := &PixelScaleTag{}
	err = scaleTag.Decode(scaleVal.Bytes(), scaleVal.Order())
	if err != nil {
		return nil, err
	}

	tieField := ifds[0].GetField(ModelTiepointTag)
	tieVal := tieField.Value()
	tieTag := &TiePointTag{}
	err = tieTag.Decode(tieVal.Bytes(), tieVal.Order())
	if err != nil {
		return nil, err
	}

	if len(tieTag.TiePoints) != 1 {
		// Being lazy
		return nil, fmt.Errorf("can't handle more than one tie point")
	} else if tieTag.TiePoints[0].PixX != 0 || tieTag.TiePoints[0].PixY != 0 {
		// Being lazy
		return nil, fmt.Errorf("tie point must be at 0,0")
	}

	// Technically I don't think we're guaranteed that the units are
	// metres, but for OS data it should be
	widthInMetres := float64(img.Bounds().Dx()) * scaleTag.ScaleX
	heightInMetres := float64(img.Bounds().Dy()) * scaleTag.ScaleY

	if widthInMetres != heightInMetres {
		return nil, fmt.Errorf("tiles must be square")
	}

	if math.Floor(widthInMetres) != widthInMetres || math.Floor(heightInMetres) != heightInMetres {
		return nil, fmt.Errorf("tile size must be whole number of metres")
	}


	tile := &Tile{
		width: osgrid.Distance(widthInMetres) * osgrid.Metre,
		height: osgrid.Distance(heightInMetres) * osgrid.Metre,
		// FIXME: May need non-integer precision
		precision: 1,
		scaleX: scaleTag.ScaleX,
		scaleY: scaleTag.ScaleY,
		image: img,
	}

	// Tie point position is top-left, so subtract height
	tile.bottomLeft, err = osgrid.Origin().Add(osgrid.Distance(tieTag.TiePoints[0].ModelX),
				osgrid.Distance(tieTag.TiePoints[0].ModelY) - tile.height)
	if err != nil {
		return nil, err
	}

	return tile, nil
}

func (d *Database) findTile(ref osgrid.GridRef) (string, error) {
	ref.Align(d.tileSize)

	dir, err := ioutil.ReadDir(d.path)
	if err != nil {
		return "", err
	}

	// It's not clear what the file structure would be for the full dataset,
	// but the single region I've got is flat.
	// I'll download the full set and sort this out once it works for a
	// single region
	for _, dirEntry := range dir {
		fileExp := fmt.Sprintf("(?i)%s%s", ref.Tile(), ref.Digits())
		fileRe := regexp.MustCompile(fileExp)
		if fileRe.MatchString(dirEntry.Name()) {
			ret := filepath.Join(d.path, dirEntry.Name())
			return ret, nil
		}
	}

	return "", fmt.Errorf("Tile %s not found", ref)
}

func findOldest(cache []tileCacheEntry) int {
	oldest := 0
	idx := 0

	for i, entry := range cache {
		if entry.timestamp <= oldest {
			oldest = entry.timestamp
			idx = i
		}
	}

	return idx
}

func (d *Database) hit(slot int) *Tile {
	tile := d.tileCache[slot].tile

	fmt.Printf("Hit %d -> %s\n", slot, tile.String())

	d.tileCache[slot].timestamp = d.timestamp
	return tile
}

func (d *Database) readAllocate(path string) (*Tile, int, error) {
	tile, err := OpenTile(path)
	if err != nil {
		return nil, -1, err
	}

	slot := findOldest(d.tileCache)

	if d.tileCache[slot].tile != nil {
		key := d.tileCache[slot].tile.String()
		evict := d.tileMap[key]
		evict.slot = -1
		d.tileMap[key] = evict
	}

	fmt.Printf("Allocate %s -> %d (%d)\n", tile.String(), slot, d.timestamp)

	d.tileCache[slot].timestamp = d.timestamp
	d.tileCache[slot].tile = tile

	return tile, slot, nil
}

func (d *Database) GetTile(ref osgrid.GridRef) (*Tile, error) {
	var tile *Tile
	var path string
	var err error

	ref = ref.Align(d.tileSize)
	key := ref.String()

	d.timestamp++

	slot := -1
	entry, ok := d.tileMap[key]
	if ok {
		path = entry.path
		if entry.slot >= 0 {
			// Cache hit
			slot = entry.slot
			tile = d.hit(slot)
		}
	} else {
		path, err = d.findTile(ref)
		if err != nil {
			return nil, err
		}
	}

	if slot < 0 {
		tile, slot, err = d.readAllocate(path)
		if err != nil {
			return nil, err
		}
	}

	d.tileMap[key] = tileMapEntry{
		path: path,
		slot: slot,
	}

	return tile, nil
}

func (d *Database) Precision() osgrid.Distance {
	return d.precision
}

func OpenDatabase(path string, tileSize osgrid.Distance) (*Database, error) {
	datapath := filepath.Join(path, "data")

	fi, err := os.Stat(datapath)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("%s should be a directory", datapath)
	}

	d := &Database{
		path: datapath,
		tileSize: tileSize,
		tileMap: make(map[string]tileMapEntry),
		tileCache: make([]tileCacheEntry, 16),
	}

	// We assume that London is available in the data-set
	// TODO: That might be a bad assumption, but good enough for now
	/*
	tq28, _ := osgrid.ParseGridRef("TQ 28")

	tile, err := d.GetTile(tq28)
	if err != nil {
		return nil, err
	}

	if tile.width != tileSize || tile.height != tileSize {
		return nil, fmt.Errorf("Specified tileSize (%d) doesn't match data (%d)", tileSize, tile.width)
	}

	d.precision = tile.Precision()
	*/
	// FIXME!
	d.precision = 1

	return d, nil
}
