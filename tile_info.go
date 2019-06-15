package main

import "encoding/xml"

type PyramidLevel struct {
	Text         string `xml:",chardata"`
	NumTilesX    int `xml:"num_tiles_x,attr"`
	NumTilesY    int `xml:"num_tiles_y,attr"`
	InverseScale string `xml:"inverse_scale,attr"`
	EmptyPelsX   string `xml:"empty_pels_x,attr"`
	EmptyPelsY   string `xml:"empty_pels_y,attr"`
}

type TileInfo struct {
	XMLName            xml.Name       `xml:"TileInfo"`
	Text               string         `xml:",chardata"`
	TileWidth          string         `xml:"tile_width,attr"`
	TileHeight         string         `xml:"tile_height,attr"`
	FullPyramidDepth   string         `xml:"full_pyramid_depth,attr"`
	Origin             string         `xml:"origin,attr"`
	Timestamp          string         `xml:"timestamp,attr"`
	TilerVersionNumber string         `xml:"tiler_version_number,attr"`
	ImageWidth         string         `xml:"image_width,attr"`
	ImageHeight        string         `xml:"image_height,attr"`
	PyramidLevel       []PyramidLevel `xml:"pyramid_level"`
}
