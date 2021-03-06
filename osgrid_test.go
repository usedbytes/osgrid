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
			tile: "ST",
			easting: 0,
			northing: 0,
		},
	},
	{
		str: "ST 001000",
		ref: GridRef{
			tile: "ST",
			easting: 100,
			northing: 0,
		},
	},
	{
		str: "ST23",
		ref: GridRef{
			tile: "ST",
			easting: 20000,
			northing: 30000,
		},
	},
	{
		str: "ST 001002",
		ref: GridRef{
			tile: "ST",
			easting: 100,
			northing: 200,
		},
	},
	{
		str: "ST 0000100002",
		ref: GridRef{
			tile: "ST",
			easting: 1,
			northing: 2,
		},
	},
	{
		str: "OG1256",
		ref: GridRef{
			tile: "OG",
			easting: 12000,
			northing: 56000,
		},
	},
	{
		str: "TL123456",
		ref: GridRef{
			tile: "TL",
			easting: 12300,
			northing: 45600,
		},
	},
	{
		str: "NT 5432 9876",
		ref: GridRef{
			tile: "NT",
			easting: 54320,
			northing: 98760,
		},
	},
	{
		str: "NT 5432198765",
		ref: GridRef{
			tile: "NT",
			easting: 54321,
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
	non GridRef
	aligned GridRef
	to Distance
}

var roundTests []roundTest = []roundTest{
	{
		non: GridRef{
			tile: "ST",
			easting: 21000,
			northing: 34000,
		},
		aligned: GridRef{
			tile: "ST",
			easting: 20000,
			northing: 30000,
		},
		to: 10 * Kilometre,
	},
	{
		non: GridRef{
			tile: "ST",
			easting: 21000,
			northing: 34000,
		},
		aligned: GridRef{
			tile: "ST",
			easting: 21000,
			northing: 34000,
		},
		to: 1 * Kilometre,
	},
	{
		non: GridRef{
			tile: "NT",
			easting: 54321,
			northing: 98765,
		},
		aligned: GridRef{
			tile: "NT",
			easting: 54320,
			northing: 98760,
		},
		to: 10 * Metre,
	},
	{
		non: GridRef{
			tile: "NT",
			easting: 54321,
			northing: 98765,
		},
		aligned: GridRef{
			tile: "NT",
			easting: 54321,
			northing: 98765,
		},
		to: 1 * Metre,
	},
	{
		non: GridRef{
			tile: "NT",
			easting: 54321,
			northing: 98765,
		},
		aligned: GridRef{
			tile: "NT",
			easting: 54321,
			northing: 98765,
		},
		to: 0 * Metre,
	},
	{
		non: GridRef{
			tile: "NT",
			easting: 54321,
			northing: 98765,
		},
		aligned: GridRef{
			tile: "NT",
			easting: 54320,
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
	a, b string
	east, north Distance
}

var addTests []addTest = []addTest{
	{
		a: "SV 00",
		b: "SW 00",
		east  : tileSize,
		north : 0,
	},
	{
		a: "SV 0000100002",
		b: "SV 00",
		east  : -1 * Metre,
		north : -2 * Metre,
	},
	{
		a: "SV 00",
		b: "SV 50",
		east  : tileSize / 2,
		north : 0,
	},
	{
		a: "NL 00",
		b: "NN 00",
		east  : tileSize * 2,
		north : 0,
	},
	{
		a: "NL 00",
		b: "OL 00",
		east  : tileSize * 5,
		north : 0,
	},
	{
		a: "SO 00",
		b: "SJ 00",
		east  : 0,
		north : tileSize,
	},
	{
		a: "SO 00",
		b: "NO 00",
		east  : 0,
		north : 5 * tileSize,
	},
	{
		a: "SN 1005",
		b: "OF 050055",
		east  : 3 * tileSize - 5 * Kilometre,
		north : 6 * tileSize + 500 * Metre,
	},
	{
		a: "HZ 00",
		b: "OA 00",
		east  : tileSize,
		north : -tileSize,
	},
	{
		a: "NG 00",
		b: "GZ 00",
		east  : -2 * tileSize,
		north : 2 * tileSize,
	},
	{
		a: "NR 00",
		b: "RE 00",
		east  : -2 * tileSize,
		north : -2 * tileSize,
	},
	{
		a: "NJ 00",
		b: "JV 00",
		east  : 2 * tileSize,
		north : 2 * tileSize,
	},
	{
		a: "NT 00",
		b: "TA 00",
		east  : 2 * tileSize,
		north : -2 * tileSize,
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

func TestExample(t *testing.T) {
	summit, _ := ParseGridRef("SH 60986 54375")
	point, _ := summit.Add(300 * Metre, 2 * Kilometre)
	fmt.Println(point.String())
}

/*
func TestDrawGrid(t *testing.T) {
	origin, _ := ParseGridRef("SV 00")
	for north := Distance(12 * tileSize); north >= 0; north -= tileSize {
		for east := 0 * Metre; east < 7 * tileSize; east += tileSize {
			val, _ := origin.Add(east, north)
			fmt.Printf("%s ", val.String()[:3])
		}
		fmt.Println("")
	}
}
*/
