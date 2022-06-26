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
			tile:     "ST",
			easting:  0,
			northing: 0,
		},
	},
	{
		str: "ST 001000",
		ref: GridRef{
			tile:     "ST",
			easting:  100,
			northing: 0,
		},
	},
	{
		str: "ST23",
		ref: GridRef{
			tile:     "ST",
			easting:  20000,
			northing: 30000,
		},
	},
	{
		str: "ST 001002",
		ref: GridRef{
			tile:     "ST",
			easting:  100,
			northing: 200,
		},
	},
	{
		str: "ST 0000100002",
		ref: GridRef{
			tile:     "ST",
			easting:  1,
			northing: 2,
		},
	},
	{
		str: "OG1256",
		ref: GridRef{
			tile:     "OG",
			easting:  12000,
			northing: 56000,
		},
	},
	{
		str: "TL123456",
		ref: GridRef{
			tile:     "TL",
			easting:  12300,
			northing: 45600,
		},
	},
	{
		str: "NT 5432 9876",
		ref: GridRef{
			tile:     "NT",
			easting:  54320,
			northing: 98760,
		},
	},
	{
		str: "NT 5432198765",
		ref: GridRef{
			tile:     "NT",
			easting:  54321,
			northing: 98765,
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
	non     GridRef
	aligned GridRef
	to      Distance
}

var roundTests []roundTest = []roundTest{
	{
		non: GridRef{
			tile:     "ST",
			easting:  21000,
			northing: 34000,
		},
		aligned: GridRef{
			tile:     "ST",
			easting:  20000,
			northing: 30000,
		},
		to: 10 * Kilometre,
	},
	{
		non: GridRef{
			tile:     "ST",
			easting:  21000,
			northing: 34000,
		},
		aligned: GridRef{
			tile:     "ST",
			easting:  21000,
			northing: 34000,
		},
		to: 1 * Kilometre,
	},
	{
		non: GridRef{
			tile:     "NT",
			easting:  54321,
			northing: 98765,
		},
		aligned: GridRef{
			tile:     "NT",
			easting:  54320,
			northing: 98760,
		},
		to: 10 * Metre,
	},
	{
		non: GridRef{
			tile:     "NT",
			easting:  54321,
			northing: 98765,
		},
		aligned: GridRef{
			tile:     "NT",
			easting:  54321,
			northing: 98765,
		},
		to: 1 * Metre,
	},
	{
		non: GridRef{
			tile:     "NT",
			easting:  54321,
			northing: 98765,
		},
		aligned: GridRef{
			tile:     "NT",
			easting:  54321,
			northing: 98765,
		},
		to: 0 * Metre,
	},
	{
		non: GridRef{
			tile:     "NT",
			easting:  54321,
			northing: 98765,
		},
		aligned: GridRef{
			tile:     "NT",
			easting:  54320,
			northing: 98764,
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

type addTest struct {
	a, b        string
	east, north Distance
}

var addTests []addTest = []addTest{
	{
		a:     "SV 00",
		b:     "SW 00",
		east:  tileSize,
		north: 0,
	},
	{
		a:     "SV 0000100002",
		b:     "SV 00",
		east:  -1 * Metre,
		north: -2 * Metre,
	},
	{
		a:     "SV 00",
		b:     "SV 50",
		east:  tileSize / 2,
		north: 0,
	},
	{
		a:     "NL 00",
		b:     "NN 00",
		east:  tileSize * 2,
		north: 0,
	},
	{
		a:     "NL 00",
		b:     "OL 00",
		east:  tileSize * 5,
		north: 0,
	},
	{
		a:     "SO 00",
		b:     "SJ 00",
		east:  0,
		north: tileSize,
	},
	{
		a:     "SO 00",
		b:     "NO 00",
		east:  0,
		north: 5 * tileSize,
	},
	{
		a:     "SN 1005",
		b:     "OF 050055",
		east:  3*tileSize - 5*Kilometre,
		north: 6*tileSize + 500*Metre,
	},
	{
		a:     "HZ 00",
		b:     "OA 00",
		east:  tileSize,
		north: -tileSize,
	},
	{
		a:     "NG 00",
		b:     "GZ 00",
		east:  -2 * tileSize,
		north: 2 * tileSize,
	},
	{
		a:     "NR 00",
		b:     "RE 00",
		east:  -2 * tileSize,
		north: -2 * tileSize,
	},
	{
		a:     "NJ 00",
		b:     "JV 00",
		east:  2 * tileSize,
		north: 2 * tileSize,
	},
	{
		a:     "NT 00",
		b:     "TA 00",
		east:  2 * tileSize,
		north: -2 * tileSize,
	},
}

func TestAdd(t *testing.T) {
	for i, test := range addTests {
		// A to B
		a, err := ParseGridRef(test.a)
		if err != nil {
			t.Error(i, err)
		}
		result, err := a.Add(test.east, test.north)
		if err != nil {
			t.Error(i, err)
		}
		if result.String() != test.b {
			t.Errorf("%d Got: %s, Expected: %s\n", i, result, test.b)
		}

		// B to A
		b, err := ParseGridRef(test.b)
		if err != nil {
			t.Error(i, err)
		}
		result, err = b.Add(-test.east, -test.north)
		if err != nil {
			t.Error(i, err)
		}
		if result.String() != test.a {
			t.Errorf("%d Got: %s, Expected: %s\n", i, result, test.a)
		}
	}
}

func TestSub(t *testing.T) {
	for i, test := range addTests {
		a, err := ParseGridRef(test.a)
		if err != nil {
			t.Error(i, err)
		}

		b, err := ParseGridRef(test.b)
		if err != nil {
			t.Error(i, err)
		}

		// B sub A
		e, n := b.Sub(a)

		if e != test.east || n != test.north {
			t.Errorf("%d b.Sub(a) got: %v,%v, expected: %v,%v", i, e, n, test.east, test.north)
		}

		// A sub B
		e, n = a.Sub(b)

		if e != -test.east || n != -test.north {
			t.Errorf("%d a.Sub(b) got: %v,%v, expected: %v,%v", i, e, n, -test.east, -test.north)
		}
	}
}

type absTest struct {
	str         string
	absEasting  Distance
	absNorthing Distance
}

var absTests []absTest = []absTest{
	{
		"SV 0 0",
		0,
		0,
	},
	{
		"SV 3 1",
		30 * Kilometre,
		10 * Kilometre,
	},
	{
		"SV 01 02",
		1 * Kilometre,
		2 * Kilometre,
	},
	{
		"SV 007 009",
		700 * Metre,
		900 * Metre,
	},
	{
		"SV 0003 0001",
		30 * Metre,
		10 * Metre,
	},
	{
		"SV 00001 00001",
		1 * Metre,
		1 * Metre,
	},
	{
		"TQ 389 773",
		5*100*Kilometre + 3*10*Kilometre + 8*1*Kilometre + 9*100*Metre,
		1*100*Kilometre + 7*10*Kilometre + 7*1*Kilometre + 3*100*Metre,
	},
	{
		"RV 0 0",
		-1 * 500 * Kilometre,
		0,
	},
	{
		"WV 0 0",
		-1 * 500 * Kilometre,
		-1 * 500 * Kilometre,
	},
}

func TestAbs(t *testing.T) {
	for i, test := range absTests {
		ref, err := ParseGridRef(test.str)
		if err != nil {
			t.Error(err)
			continue
		}

		ae, an := ref.AbsEasting(), ref.AbsNorthing()

		if ae != test.absEasting || an != test.absNorthing {
			t.Errorf("%d %s got: %v,%v, expected: %v,%v", i, test.str, ae, an, test.absEasting, test.absNorthing)
		}
	}
}
