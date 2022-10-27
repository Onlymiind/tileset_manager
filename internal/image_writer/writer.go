package image_writer

import (
	"image"
	"image/color"

	"github.com/Onlymiind/tileset_manager/internal/common"
	"github.com/Onlymiind/tileset_manager/internal/file_manager"
)

type outPalette []color.Color

func MakePalette(palette []color.Color) outPalette {
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

func WriteTileData(tileData common.Tiles, palette [4]color.Color) *image.Paletted {
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

func writeMetatileTile(cache *file_manager.TileCache, tileset *common.Metatiles, img *image.Paletted, index uint8, palette outPalette, x, y int) {
	refIt := tileset.Refs.Find(common.TileRef{Range: common.IndexRange{Start: index, End: index}})
	if refIt == nil {
		return
	}
	ref := refIt.GetValue()
	if len(ref.File) == 0 {
		return
	}

	tile, err := cache.GetTile(ref.File, ref.Offset+(index-ref.Range.Start))
	if err != nil {
		return
	}

	writeTileToImage(img, palette, tile, x, y)

}

func WriteMetatileData(cache *file_manager.TileCache, tileset *common.Metatiles) *image.Paletted {
	width := common.OutTilesPerRow
	if len(tileset.Metatiles) < width {
		width = len(tileset.Metatiles)
	}

	height := len(tileset.Metatiles) / width
	if len(tileset.Metatiles)%width != 0 {
		height++
	}

	actualPalette := make(outPalette, 4)
	copy(actualPalette, tileset.Palette[:])
	if tileset.AbsentTiles.Size() != 0 {
		actualPalette = MakePalette(tileset.Palette)
	}

	img := image.NewPaletted(image.Rect(0, 0, width*common.MetatileSizePx, height*common.MetatileSizePx),
		[]color.Color(actualPalette))

	x, y := 0, 0

	for _, mtile := range tileset.Metatiles {
		writeMetatileTile(cache, tileset, img, mtile.TopLeft, actualPalette, x, y)
		writeMetatileTile(cache, tileset, img, mtile.TopRight, actualPalette, x+common.TileSizePx, y)
		writeMetatileTile(cache, tileset, img, mtile.BottomLeft, actualPalette, x, y+common.TileSizePx)
		writeMetatileTile(cache, tileset, img, mtile.BottomRight, actualPalette, x+common.TileSizePx, y+common.TileSizePx)
		x += common.MetatileSizePx
		if x >= width*common.MetatileSizePx {
			x %= width * common.MetatileSizePx
			y += common.MetatileSizePx
		}
	}

	return img
}
