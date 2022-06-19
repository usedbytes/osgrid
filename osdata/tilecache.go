package osdata

import (
	"fmt"
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
	stats     stats
}

type stats struct {
	hits int
	misses int
	evictions int
	allocations int
}

func NewCache(nslots int) *Cache {
	return &Cache{
		nslots: nslots,
		slots:  make([]slot, nslots),
		cache:  make(map[osgrid.GridRef]*entry),
	}
}

func (c *Cache) read(ref osgrid.GridRef) (Tile, bool) {
	if entry, ok := c.cache[ref]; ok {
		c.timestamp++
		c.slots[entry.slot].ts = c.timestamp
		return entry.tile, ok
	}

	return nil, false
}

func (c *Cache) Read(ref osgrid.GridRef) (Tile, bool) {
	entry, ok := c.read(ref)
	if ok {
		c.stats.hits++
	} else {
		c.stats.misses++
	}

	return entry, ok
}

func (c *Cache) findSlot() int {
	oldest := c.timestamp
	idx := 0

	for i, slot := range c.slots {
		if slot.ref.Tile() == "" {
			// Always prefer an empty slot
			idx = i
			return idx
		}

		if slot.ts <= oldest {
			oldest = slot.ts
			idx = i
		}
	}

	c.stats.evictions++

	return idx
}

func (c *Cache) Allocate(tile Tile) {
	ref := tile.BottomLeft()
	// Check we don't already have it
	if _, ok := c.read(ref); ok {
		return
	}

	idx := c.findSlot()
	slot := &c.slots[idx]

	delete(c.cache, slot.ref)
	c.cache[ref] = &entry{
		slot: idx,
		tile: tile,
	}

	c.timestamp++
	slot.ts = c.timestamp
	slot.ref = ref

	c.stats.allocations++
}

func (c *Cache) dump() string {
	s := ""
	s += fmt.Sprintf("Num slots: %d\n", c.nslots)
	s += fmt.Sprintf("Timestamp: %d\n", c.timestamp)
	s += fmt.Sprintf("Slots:\n")
	for i, v := range c.slots {
		s += fmt.Sprintf("\t%v: %+v\n", i, v)
	}
	for k, v := range c.cache {
		s += fmt.Sprintf("\t%v: %+v\n", k, v)
	}

	return s
}

func (c *Cache) DumpStats() string {
	return fmt.Sprintf("Hits: %d, Miss: %d, Allocations: %d, Evictions: %d",
		c.stats.hits, c.stats.misses, c.stats.allocations, c.stats.evictions)
}
