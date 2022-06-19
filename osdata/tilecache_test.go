package osdata

import (
	"testing"

	"github.com/usedbytes/osgrid"
)

type TestTile struct {
	ref osgrid.GridRef
}

func (tt *TestTile) Width() osgrid.Distance {
	return 10 * osgrid.Kilometre
}

func (tt *TestTile) Height() osgrid.Distance {
	return 10 * osgrid.Kilometre
}

func (tt *TestTile) Precision() osgrid.Distance {
	return 10 * osgrid.Metre
}

func (tt *TestTile) BottomLeft() (osgrid.GridRef) {
	return tt.ref
}

func (tt *TestTile) String() string {
	return tt.ref.String()
}

func NewTestTile(ref osgrid.GridRef) Tile {
	return &TestTile{
		ref: ref,
	}
}

func TestCacheEmpty(t *testing.T) {
	c := NewCache(1)

	ref, err := osgrid.ParseGridRef("SH 60 54")
	if err != nil {
		panic(err)
	}

	_, ok := c.Read(ref)
	if ok {
		t.Fatal("should have missed for non-allocated ref")
	}
}

func TestCacheSingleAllocate(t *testing.T) {
	c := NewCache(1)

	ref, err := osgrid.ParseGridRef("SH 60 54")
	if err != nil {
		panic(err)
	}

	_, ok := c.Read(ref)
	if ok {
		t.Fatal("should have missed for non-allocated ref")
	}

	tt := NewTestTile(ref)
	c.Allocate(tt)

	if len(c.cache) != 1 {
		t.Fatal("should have one occupied slot")
	}

	if c.cache[ref].slot != 0 {
		t.Fatal("slot 0 should have been allocated")
	}

	if c.cache[ref].tile != tt {
		t.Fatal("entry tile doesn't match")
	}

	if c.slots[0].ref != ref {
		t.Fatalf("slot ref not expected. want: %v, got: %v", ref, c.slots[0].ref)
	}
}

func TestCacheSingleHit(t *testing.T) {
	c := NewCache(1)

	ref, err := osgrid.ParseGridRef("SH 60 54")
	if err != nil {
		panic(err)
	}

	tt := NewTestTile(ref)
	c.Allocate(tt)

	_, ok := c.Read(ref)
	if !ok {
		t.Fatal("should have hit")
	}
}

func TestCacheSingleEvict(t *testing.T) {
	c := NewCache(1)

	ref1, err := osgrid.ParseGridRef("SH 60 54")
	if err != nil {
		panic(err)
	}

	tt1 := NewTestTile(ref1)
	c.Allocate(tt1)

	ref2, err := osgrid.ParseGridRef("SG 60 54")
	if err != nil {
		panic(err)
	}

	_, ok := c.Read(ref2)
	if ok {
		t.Fatal("should have missed for non-allocated ref2")
	}

	tt2 := NewTestTile(ref2)
	c.Allocate(tt2)

	_, ok = c.Read(ref1)
	if ok {
		t.Fatal("should have missed for evicted ref1")
	}

	_, ok = c.Read(ref2)
	if !ok {
		t.Fatal("should have hit for ref2")
	}

	if len(c.cache) != 1 {
		t.Fatal("should have one slot")
	}

	if c.cache[ref2].slot != 0 {
		t.Fatal("slot 0 should have been replaced")
	}

	if c.cache[ref2].tile != tt2 {
		t.Fatal("entry tile doesn't match")
	}

	if c.slots[0].ref != ref2 {
		t.Fatalf("slot ref not expected. want: %v, got: %v", ref2, c.slots[0].ref)
	}
}

func TestCacheMultiEvict(t *testing.T) {
	const cacheSize int = 3
	c := NewCache(cacheSize)

	tiles := make([]Tile, cacheSize * 2)
	ref := osgrid.Origin()

	var err error
	for i := range tiles {
		ref, err = ref.Add(10 * osgrid.Kilometre, 0)
		if err != nil {
			panic(err)
		}
		tiles[i] = NewTestTile(ref)

		c.Allocate(tiles[i])
		if i > cacheSize {
			_, ok := c.Read(tiles[i - cacheSize - 1].BottomLeft())
			if ok {
				t.Fatalf("should have missed for evicted ref. idx %d", i)
			}
		}
	}
}

func TestCacheMultiEvictOldest(t *testing.T) {
	const cacheSize int = 3
	c := NewCache(cacheSize)

	tiles := make([]Tile, cacheSize + 1)
	ref := osgrid.Origin()

	var err error
	for i := range tiles {
		ref, err = ref.Add(10 * osgrid.Kilometre, 0)
		if err != nil {
			panic(err)
		}
		tiles[i] = NewTestTile(ref)
	}

	for _, t := range tiles[:cacheSize] {
		c.Allocate(t)
	}

	_, ok := c.Read(tiles[0].BottomLeft())
	if !ok {
		t.Fatalf("should have hit for oldest entry")
	}

	c.Allocate(tiles[cacheSize])

	_, ok = c.Read(tiles[0].BottomLeft())
	if !ok {
		t.Fatalf("refreshed entry should not have been evicted")
	}

	_, ok = c.Read(tiles[1].BottomLeft())
	if ok {
		t.Fatalf("second oldest entry should have been evicted")
	}
}
