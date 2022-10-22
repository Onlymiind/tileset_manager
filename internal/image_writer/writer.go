package image_writer

import (
	"image"
	"image/color"

	"github.com/Onlymiind/tileset_manager/internal/common"
	"github.com/Onlymiind/tileset_manager/proto"
)

type outPalette []color.Color

func MakePalette(palette [4]color.Color) outPalette {
	actualPalette := make(outPalette, 0, 5)
	actualPalette = append(actualPalette, color.Transparent)
	actualPalette = append(actualPalette, palette[:]...)
	return actualPalette
}

func (p outPalette) getColorIndex(rawIndex uint8) uint8 {
	if len(p) == 4 {
		return rawIndex
	}
	return rawIndex + 1
}

func writeTileToImage(image *image.Paletted, palette outPalette, tile []byte, x, y int) {
	if len(tile) != common.BitsPerTile {
		return
	}

	for row := 0; row < common.TileSizePx; row++ {
		for column := 0; column < common.TileSizePx; column++ {
			image.SetColorIndex(x+column, y+row, palette.getColorIndex(tile[row*common.TileSizePx+column]))
		}
	}
}

func WriteTileData(tileData [][]byte, palette [4]color.Color) *image.Paletted {
	width := common.OutTilesPerRow
	if len(tileData) < width {
		width = len(tileData)
	}

	height := len(tileData) / width
	if len(tileData)%width != 0 {
		height++
	}

	img := image.NewPaletted(image.Rect(0, 0, width*common.TileSizePx, height*common.TileSizePx),
		[]color.Color(palette[:]))
	x, y := 0, 0
	for _, tile := range tileData {
		writeTileToImage(img, palette[:], tile, x, y)
		x += common.TileSizePx
		if x >= width*common.TileSizePx {
			x %= width * common.TileSizePx
			y += common.TileSizePx
		}
	}

	return img
}

func WriteMetatileData(tileset *proto.Tileset, palette [4]color.Color) *image.Paletted {
	width := common.OutTilesPerRow
	if len(tileset.Metatiles) < width {
		width = len(tileset.Metatiles)
	}

	height := len(tileset.Metatiles) / width
	if len(tileset.Metatiles)%width != 0 {
		height++
	}

	actualPalette := make(outPalette, 4)
	copy(actualPalette, palette[:])
	needTransparent := false
	for _, mtile := range tileset.Metatiles {
		_, ok1 := tileset.TileData[mtile.TopLeft]
		_, ok2 := tileset.TileData[mtile.TopRight]
		_, ok3 := tileset.TileData[mtile.BottomLeft]
		_, ok4 := tileset.TileData[mtile.BottomRight]
		needTransparent = !(ok1 && ok2 && ok3 && ok4)
		if needTransparent {
			break
		}
	}

	if needTransparent {
		actualPalette = MakePalette(palette)
	}

	img := image.NewPaletted(image.Rect(0, 0, width*common.MetatileSizePx, height*common.MetatileSizePx),
		[]color.Color(actualPalette))

	x, y := 0, 0

	for _, mtile := range tileset.Metatiles {
		writeTileToImage(img, actualPalette, tileset.TileData[mtile.TopLeft], x, y)
		writeTileToImage(img, actualPalette, tileset.TileData[mtile.TopRight], x+common.TileSizePx, y)
		writeTileToImage(img, actualPalette, tileset.TileData[mtile.BottomLeft], x, y+common.TileSizePx)
		writeTileToImage(img, actualPalette, tileset.TileData[mtile.BottomRight], x+common.TileSizePx, y+common.TileSizePx)
		x += common.MetatileSizePx
		if x >= width*common.MetatileSizePx {
			x %= width * common.MetatileSizePx
			y += common.MetatileSizePx
		}
	}

	return img
}
