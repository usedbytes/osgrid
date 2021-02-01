Ordnance Survey - The National Grid
===================================

The base `osgrid` package provides library functions for working with the
[Ordnance Survey National Grid](https://www.ordnancesurvey.co.uk/resources/maps-and-geographic-resources/the-national-grid.html)
mapping coordinate system used in the United Kingdom.

`osgrid` can parse, print and do arithmetic on OS "grid references"

For example, the OS grid reference of the summit of Snowdon is `SH 60986 54375`.
We can find the OS grid reference for the point 300 metres East and 2 kilometres
North like so:

```
	summit, _ := osgrid.ParseGridRef("SH 60986 54375")
	point, _ := summit.Add(300 * osgrid.Metre, 2 * osgrid.Kilometre)
	fmt.Println(point.String())
```

`terrain50`
-----------

The Ordnance Survey makes much of their data available for download for free
from [their website](https://osdatahub.os.uk/downloads/open).

One such dataset is the [_Terrain 50_](https://osdatahub.os.uk/downloads/open/Terrain50)
which provides elevation data, with 0.1 m vertical and 50 m horizontal resolution,
for the whole of the United Kingdom.

The `terrain50` package provides a way to work with this dataset, giving easy
access to elevation data by grid-reference.

The `terrain50.Database` object represents the data set. To use it, simply
download the _Terrain 50_ "ASCII Grid" dataset, and extract it to a location. Use
this location to open the database. The path should be to the directory _which
contains the `data` directory_:

```
	db, _ := terrain50.OpenDatabase("/path/to/terrain_50_data", 10 * osgrid.Kilometre)

```

If all goes well, you can now start asking for elevation information for a point:
```
	summit, _ := osgrid.ParseGridRef("SH 60986 54375")
	elevation, _ := db.GetData(summit)
	fmt.Printf("Snowdon's summit is at %f m", elevation)
```

The `terrain50.Database` implements a tile cache (with 16 entries by default),
storing the data for the 16 most recently used tiles, so that queries which are
geographically close to each other are fast, and to ensure memory usage doesn't
grow unbounded.

`surface`
---------

`surface` is a simple executable tool which generates [OpenSCAD](http://www.openscad.org/)
renders of the _Terrain 50_ data.

It can directly generate an OpenSCAD render for a given grid-reference and radius:

```
$ ./surface -d /path/to/terrain_50_data -g SH6098654375 --xyscale 1:100000 --zscale 1:50000 --radius 5000 --output out.dat --scad out.scad
```

![Snowdon Summit (5 km radius)](surface.png)

> Image above contains OS data © Crown Copyright (2021), used under the
> [Open Government License](http://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/)
