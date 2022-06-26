package osgrid

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Distance int

const (
	Metre     Distance = 1
	Kilometre          = 1000
	tileSize           = 100 * Kilometre
)

type GridRef struct {
	tile              string
	easting, northing Distance
}

func Origin() GridRef {
	return GridRef{"SV", 0, 0}
}

const gridChars string = "ABCDEFGHJKLMNOPQRSTUVWXYZ"
const digits string = "0123456789"

func validGridChar(c rune) bool {
	return strings.ContainsRune(gridChars, c)
}

func validDigit(c rune) bool {
	return strings.ContainsRune(digits, c)
}

func isNumeric(str string) bool {
	for _, c := range str {
		if !validDigit(c) {
			return false
		}
	}
	return true
}

func ParseGridRef(str string) (GridRef, error) {
	str = strings.Replace(strings.ToUpper(str), " ", "", -1)

	square := str[0:2]
	if !validGridChar(rune(square[0])) || !validGridChar(rune(square[1])) {
		return GridRef{}, fmt.Errorf("Invalid square '%s'", square)
	}

	numeric := str[2:]
	if !isNumeric(numeric) {
		return GridRef{}, fmt.Errorf("Invalid digits '%s'", numeric)
	}
	if len(numeric)%2 != 0 {
		return GridRef{}, fmt.Errorf("Need an even number of digits '%s'", numeric)
	}

	diff := 5 - (len(numeric) / 2)
	mult := math.Pow10(diff)

	easting, err := strconv.ParseFloat(numeric[:len(numeric)/2], 64)
	if err != nil {
		return GridRef{}, fmt.Errorf("Couldn't parse easting '%s'", numeric[:len(numeric)/2])
	}
	northing, err := strconv.ParseFloat(numeric[len(numeric)/2:], 64)
	if err != nil {
		return GridRef{}, fmt.Errorf("Couldn't parse northing '%s'", numeric[len(numeric)/2:])
	}

	return GridRef{
		tile:     square,
		easting:  Distance(easting * mult),
		northing: Distance(northing * mult),
	}, nil
}

func (g GridRef) Tile() string {
	return g.tile
}

func (g GridRef) Digits() string {
	digits := 5

	easting, northing := g.easting, g.northing
	for {
		if easting%10 != 0 || northing%10 != 0 || digits == 1 {
			break
		}
		easting, northing, digits = easting/10, northing/10, digits-1
	}

	return fmt.Sprintf("%0*d%0*d", digits, easting, digits, northing)
}

func (g GridRef) String() string {
	return fmt.Sprintf("%s %s", g.Tile(), g.Digits())
}

func (g GridRef) TileEasting() Distance {
	return g.easting
}

func (g GridRef) TileNorthing() Distance {
	return g.northing
}

// Get Easting, Northing from AA00 to gridRef
func (g GridRef) distFromAA00() (Distance, Distance) {
	tile := g.Tile()

	first := strings.IndexRune(gridChars, rune(tile[0]))
	second := strings.IndexRune(gridChars, rune(tile[1]))

	southing := (Distance(first/5) * 500 * Kilometre) + (Distance(second/5) * 100 * Kilometre) - g.TileNorthing()
	easting := (Distance(first%5) * 500 * Kilometre) + (Distance(second%5) * 100 * Kilometre) + g.TileEasting()

	return easting, -southing
}

func (g GridRef) Sub(b GridRef) (Distance, Distance) {
	ae, an := g.distFromAA00()
	be, bn := b.distFromAA00()

	return ae - be, an - bn
}

func (g GridRef) AbsEasting() Distance {
	ae, _ := g.distFromAA00()
	be, _ := Origin().distFromAA00()

	return ae - be
}

func (g GridRef) AbsNorthing() Distance {
	_, an := g.distFromAA00()
	_, bn := Origin().distFromAA00()

	return an - bn
}

func (g GridRef) Align(to Distance) GridRef {
	if to == 0 {
		to = 1
	}

	return GridRef{
		tile:     g.tile,
		easting:  (g.easting / to) * to,
		northing: (g.northing / to) * to,
	}
}

func horizontalAdd(idx, add int) (int, bool) {
	idx += add

	rem := idx % 5
	if rem < 0 {
		rem += 5
	}

	if add > 0 && rem == 0 {
		idx -= 5
		return idx, true
	} else if add < 0 && rem == 4 {
		idx += 5
		return idx, true
	}

	return idx, false
}

func moveEastWest(tile string, dir int) (string, error) {
	if len(tile) != 2 {
		return "", fmt.Errorf("Invalid tile '%s'", tile)
	}

	first := strings.IndexRune(gridChars, rune(tile[0]))
	second := strings.IndexRune(gridChars, rune(tile[1]))
	if first < 0 || second < 0 {
		return "", fmt.Errorf("Invalid tile '%s'", tile)
	}

	second, carry := horizontalAdd(second, dir)
	if carry {
		first, carry = horizontalAdd(first, dir)
		if carry {
			return "", fmt.Errorf("Fell off the edge of the flat world")
		}
	}

	return string(gridChars[first]) + string(gridChars[second]), nil
}

func verticalAdd(idx, add int) (int, bool) {
	idx += add

	rem := idx % 25
	if rem < 0 {
		rem += 25
	}
	return rem, (idx >= 25 || idx < 0)
}

func moveNorthSouth(tile string, dir int) (string, error) {
	if len(tile) != 2 {
		return "", fmt.Errorf("Invalid tile '%s'", tile)
	}

	first := strings.IndexRune(gridChars, rune(tile[0]))
	second := strings.IndexRune(gridChars, rune(tile[1]))
	if first < 0 || second < 0 {
		return "", fmt.Errorf("Invalid tile '%s'", tile)
	}

	second, carry := verticalAdd(second, dir)
	if carry {
		first, carry = verticalAdd(first, dir)
		if carry {
			return "", fmt.Errorf("Fell off the edge of the flat world")
		}
	}

	return string(gridChars[first]) + string(gridChars[second]), nil
}

func (g GridRef) Add(east Distance, north Distance) (GridRef, error) {
	var err error

	g.easting += east
	g.northing += north

	for ; g.easting >= tileSize; g.easting -= tileSize {
		g.tile, err = moveEastWest(g.tile, 1)
		if err != nil {
			return GridRef{}, err
		}
	}

	for ; g.easting < 0; g.easting += tileSize {
		g.tile, err = moveEastWest(g.tile, -1)
		if err != nil {
			return GridRef{}, err
		}
	}

	for ; g.northing >= tileSize; g.northing -= tileSize {
		g.tile, err = moveNorthSouth(g.tile, -5)
		if err != nil {
			return GridRef{}, err
		}
	}

	for ; g.northing < 0; g.northing += tileSize {
		g.tile, err = moveNorthSouth(g.tile, 5)
		if err != nil {
			return GridRef{}, err
		}
	}

	return g, nil
}
