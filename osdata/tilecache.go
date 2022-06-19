package osdata

import (
	"github.com/usedbytes/osgrid"
)

type slot struct {
	ts  int
	ref osgrid.GridRef
}

type entry struct {
	slot int
	tile Tile
}

type Cache struct {
	nslots    int
	timestamp int
	slots     []slot
	cache     map[osgrid.GridRef]*entry
}

func NewCache(nslots int) *Cache {
	return &Cache{
		nslots: nslots,
		slots:  make([]slot, nslots, 0),
		cache:  make(map[osgrid.GridRef]*entry),
	}
}

func (c *Cache) Read(ref osgrid.GridRef) (Tile, bool) {
	if entry, ok := c.cache[ref]; ok {
		c.timestamp++
		c.slots[entry.slot].ts = c.timestamp
		return entry.tile, ok
	}

	return nil, false
}

func findOldest(slots []slot) int {
	oldest := 0
	idx := 0

	for i, slot := range slots {
		if slot.ts <= oldest {
			oldest = slot.ts
			idx = i
		}
	}

	return idx
}

func (c *Cache) Allocate(tile Tile) {
	ref := tile.BottomLeft()
	// Check we don't already have it
	if _, ok := c.Read(ref); ok {
		return
	}

	idx := findOldest(c.slots)
	slot := &c.slots[idx]

	delete(c.cache, slot.ref)
	c.cache[ref] = &entry{
		slot: idx,
		tile: tile,
	}

	c.timestamp++
	slot.ts = c.timestamp
	slot.ref = ref
}
