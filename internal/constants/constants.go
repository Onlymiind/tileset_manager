package constants

import "image/color"

const (
	DefaultOutDir         = "extracted"
	ExtensionTileData     = ".chr"
	ExtensionMetatileData = ".mtiles"
	OutTilesPerRow        = 16
	TileSizePx            = 8
	BitsPerTile           = TileSizePx * TileSizePx
	BytesPerTile          = TileSizePx * 2
	MetatileSizePx        = TileSizePx * 2
	AirTileID             = 0xFF
)

var AirTileData [8]byte = [8]byte{}
var DefaultPalette [4]color.Color = [4]color.Color{
	color.Black,
	color.White,
	color.Gray16{(0xffff / 4) * 2}, //Light gray
	color.Gray16{0xffff / 4},       //Dark gray
}
