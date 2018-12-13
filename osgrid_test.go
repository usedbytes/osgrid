package osgrid

import (
	"fmt"
	"strings"
	"testing"
)

type parseTest struct {
	str string
	ref GridRef
}

var parseTests []parseTest = []parseTest{
	{
		str: "ST 00",
		ref: GridRef{
			Tile: "ST",
			Easting: 0,
			Northing: 0,
		},
	},
	{
		str: "ST23",
		ref: GridRef{
			Tile: "ST",
			Easting: 20000,
			Northing: 30000,
		},
	},
	{
		str: "ST 001002",
		ref: GridRef{
			Tile: "ST",
			Easting: 100,
			Northing: 200,
		},
	},
	{
		str: "ST 0000100002",
		ref: GridRef{
			Tile: "ST",
			Easting: 1,
			Northing: 2,
		},
	},
	{
		str: "OG1256",
		ref: GridRef{
			Tile: "OG",
			Easting: 12000,
			Northing: 56000,
		},
	},
	{
		str: "TL123456",
		ref: GridRef{
			Tile: "TL",
			Easting: 12300,
			Northing: 45600,
		},
	},
	{
		str: "NT 5432 9876",
		ref: GridRef{
			Tile: "NT",
			Easting: 54320,
			Northing: 98760,
		},
	},
	{
		str: "NT 5432198765",
		ref: GridRef{
			Tile: "NT",
			Easting: 54321,
			Northing: 98765,
		},
	},
}

func TestParseGridRef(t *testing.T) {
	for _, test := range parseTests {
		got, err := ParseGridRef(test.str)
		if err != nil {
			t.Error("Parse failed.", err)
		}
		if got != test.ref {
			t.Errorf("Got: %#v (%s), Expected: %#v (%s)\n", got, got, test.ref, test.ref)
		}
	}
}

func TestFormatGridRef(t *testing.T) {
	for _, test := range parseTests {
		nospace := strings.Replace(strings.ToUpper(test.str), " ", "", -1)
		canonical := fmt.Sprintf("%s %s", nospace[:2], nospace[2:])

		str := test.ref.String()
		if str != canonical {
			t.Errorf("Got %s, Expected %s\n", str, canonical)
		}
	}
}


type roundTest struct {
	non GridRef
	aligned GridRef
	to Distance
}

var roundTests []roundTest = []roundTest{
	{
		non: GridRef{
			Tile: "ST",
			Easting: 21000,
			Northing: 34000,
		},
		aligned: GridRef{
			Tile: "ST",
			Easting: 20000,
			Northing: 30000,
		},
		to: 10 * Kilometre,
	},
	{
		non: GridRef{
			Tile: "ST",
			Easting: 21000,
			Northing: 34000,
		},
		aligned: GridRef{
			Tile: "ST",
			Easting: 21000,
			Northing: 34000,
		},
		to: 1 * Kilometre,
	},
	{
		non: GridRef{
			Tile: "NT",
			Easting: 54321,
			Northing: 98765,
		},
		aligned: GridRef{
			Tile: "NT",
			Easting: 54320,
			Northing: 98760,
		},
		to: 10 * Metre,
	},
	{
		non: GridRef{
			Tile: "NT",
			Easting: 54321,
			Northing: 98765,
		},
		aligned: GridRef{
			Tile: "NT",
			Easting: 54321,
			Northing: 98765,
		},
		to: 1 * Metre,
	},
	{
		non: GridRef{
			Tile: "NT",
			Easting: 54321,
			Northing: 98765,
		},
		aligned: GridRef{
			Tile: "NT",
			Easting: 54321,
			Northing: 98765,
		},
		to: 0 * Metre,
	},
	{
		non: GridRef{
			Tile: "NT",
			Easting: 54321,
			Northing: 98765,
		},
		aligned: GridRef{
			Tile: "NT",
			Easting: 54320,
			Northing: 98764,
		},
		to: 2 * Metre,
	},
}

func TestAlign(t *testing.T) {
	for _, test := range roundTests {
		got := test.non.Align(test.to)
		if got != test.aligned {
			t.Errorf("Got: %#v (%s), Expected: %#v (%s)\n", got, got, test.aligned, test.aligned)
		}
	}
}
