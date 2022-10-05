package image_writer

import (
	"image"
	"image/color"

	"github.com/Onlymiind/tileset_generator/internal/constants"
	"github.com/Onlymiind/tileset_generator/proto"
)

type outPalette []color.Color

func (p outPalette) getColorIndex(rawIndex uint8) uint8 {
	return rawIndex + 1
}

func writeTileToImage(image *image.Paletted, palette outPalette, tile []byte, x, y int) {
	if len(tile) != constants.BitsPerTile {
		return
	}

	for row := 0; row < constants.TileSizePx; row++ {
		for column := 0; column < constants.TileSizePx; column++ {
			image.SetColorIndex(x+column, y+row, palette.getColorIndex(tile[row*constants.TileSizePx+column]))
		}
	}
}

func WriteTileData(tileData *proto.Tiles, palette [4]color.Color) *image.Paletted {
	width := constants.OutTilesPerRow
	if len(tileData.Tiles) < width {
		width = len(tileData.Tiles)
	}

	height := len(tileData.Tiles) / width
	if len(tileData.Tiles)%width != 0 {
		height++
	}

	actualPalette := make(outPalette, 0, 5)
	actualPalette = append(actualPalette, color.Transparent)
	actualPalette = append(actualPalette, palette[:]...)

	img := image.NewPaletted(image.Rect(0, 0, width*constants.TileSizePx, height*constants.TileSizePx),
		[]color.Color(actualPalette))
	x, y := 0, 0
	for _, tile := range tileData.Tiles {
		writeTileToImage(img, actualPalette, tile, x, y)
		x += constants.TileSizePx
		if x >= constants.OutTilesPerRow*constants.TileSizePx {
			x %= constants.OutTilesPerRow * constants.TileSizePx
			y += constants.TileSizePx
		}
	}

	return img
}
