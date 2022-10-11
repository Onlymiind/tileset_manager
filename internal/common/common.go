package common

import (
	"image/color"
	"strings"
)

const (
	DefaultOutDir         = "extracted"
	ExtensionTileData     = ".chr"
	ExtensionMetatileData = ".mtile"
	ExtensionJSON         = ".json"
	OutTilesPerRow        = 16
	TileSizePx            = 8
	BitsPerTile           = TileSizePx * TileSizePx
	BytesPerTile          = TileSizePx * 2
	MetatileSizePx        = TileSizePx * 2
	AirTileID             = 0xFF
)

var AirTileData [BitsPerTile]byte = [BitsPerTile]byte{}
var DefaultPalette [4]color.Color = [4]color.Color{
	color.Black,
	color.White,
	color.Gray16{(0xffff / 4) * 2}, //Light gray
	color.Gray16{0xffff / 4},       //Dark gray
}

func ReplaceLast(src string, old string, new string) string {
	i := strings.LastIndex(src, old)
	if i < 0 {
		return src
	}
	return src[:i] + strings.ReplaceAll(src[i:], old, new)
}
