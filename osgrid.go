package osgrid

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Distance int
const (
	Metre Distance = 1
	Kilometre = 1000
)

type GridRef struct {
	tile string
	easting, northing int
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
	if len(numeric) % 2 != 0 {
		return GridRef{}, fmt.Errorf("Need an even number of digits '%s'", numeric)
	}

	diff := 5 - (len(numeric) / 2)
	mult := math.Pow10(diff)

	easting, err := strconv.ParseFloat(numeric[:len(numeric) / 2], 64)
	if err != nil {
		return GridRef{}, fmt.Errorf("Couldn't parse easting '%s'", numeric[:len(numeric) / 2])
	}
	northing, err := strconv.ParseFloat(numeric[len(numeric) / 2:], 64)
	if err != nil {
		return GridRef{}, fmt.Errorf("Couldn't parse northing '%s'", numeric[len(numeric) / 2:])
	}

	return GridRef{
		tile: square,
		easting: int(easting * mult),
		northing: int(northing * mult),
	}, nil
}

func (g GridRef) String() string {
	digits := 5

	easting, northing := g.easting, g.northing;
	for {
	    if easting % 10 != 0 || northing % 10 != 0 || digits == 1 {
		break;
	    }
	    easting, northing, digits = easting / 10, northing / 10, digits - 1
	}

	return fmt.Sprintf("%s %0*d%0*d", g.tile, digits, easting, digits, northing)
}

func (g GridRef) Align(to Distance) GridRef {
	if to == 0 {
		to = 1
	}

	return GridRef{
		tile: g.tile,
		easting: (g.easting / int(to)) * int(to),
		northing: (g.northing / int(to)) * int(to),
	}
}
