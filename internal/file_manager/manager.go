package file_manager

import (
	"errors"
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"os"
	"path"

	"github.com/Onlymiind/tileset_manager/internal/common"
	"github.com/Onlymiind/tileset_manager/internal/extractor"
	"github.com/Onlymiind/tileset_manager/internal/serializer"
	"github.com/valyala/fastjson"
)

func IsTileData(info fs.FileInfo) bool {
	return path.Ext(info.Name()) == common.ExtensionTileData && info.Mode().IsRegular()
}

func IsMetatileData(info fs.FileInfo) bool {
	return path.Ext(info.Name()) == common.ExtensionMetatileData && info.Mode().IsRegular()
}

func TileDataToImage(tileData *common.Tiles) *image.Paletted {
	width := common.OutTilesPerRow
	if len(tileData.Data) < width {
		width = len(tileData.Data)
	}

	height := len(tileData.Data) / width
	if len(tileData.Data)%width != 0 {
		height++
	}

	img := image.NewPaletted(image.Rect(0, 0, width*common.TileSizePx, height*common.TileSizePx),
		[]color.Color(tileData.Palette))
	x, y := 0, 0
	for _, tile := range tileData.Data {
		writeTileToImage(img, tileData.Palette, tile, x, y)
		x += common.TileSizePx
		if x >= width*common.TileSizePx {
			x %= width * common.TileSizePx
			y += common.TileSizePx
		}
	}

	return img
}

func ExtractTileData(filePath string) (*common.Tiles, error) {
	switch path.Ext(filePath) {
	case common.ExtensionJSON:
		return serializer.ParseTileData(filePath)
	case common.ExtensionPNG:
		//TODO
		return nil, errors.New("not implemented")
	default:
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		tileData := extractor.ExtractTileData(data)

		return tileData, nil
	}
}

func ExtractMetatileData(filePath string, tileData common.Tree[common.TileRef]) (*common.Metatiles, error) {

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	tileset := extractor.ExtractMetatileData(data, tileData)

	return tileset, nil
}

type Manager struct {
	cache tileCache
	out   common.Output
}

func NewManager(cfg *common.Config) *Manager {
	return &Manager{
		cache: newTileCache(cfg.CacheSize),
		out:   cfg.Output,
	}
}

func (m *Manager) CacheSize() common.MemorySize {
	return m.cache.getSize()
}

func (m *Manager) WritePNG(img *image.Paletted, name string, isTileData bool) error {
	imgFile, err := os.OpenFile(m.getOutPath(name, common.ExtensionPNG, isTileData), os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return common.Wrap(err, "failed to open file")
	}
	err = png.Encode(imgFile, img)
	if err != nil {
		return common.Wrap(err, "failed to encode image")
	}
	return imgFile.Close()
}

func (m *Manager) WriteJSON(json *fastjson.Value, name string, isTileData bool) error {
	err := os.WriteFile(m.getOutPath(name, common.ExtensionJSON, isTileData), json.MarshalTo(nil), 0666)
	if err != nil {
		return common.Wrap(err, "failed to write json")
	}
	return nil
}

func (m *Manager) MetatileToImage(tileset *common.Metatiles) *image.Paletted {
	width := common.OutTilesPerRow
	if len(tileset.Metatiles) < width {
		width = len(tileset.Metatiles)
	}

	height := len(tileset.Metatiles) / width
	if len(tileset.Metatiles)%width != 0 {
		height++
	}

	actualPalette := make(outPalette, len(tileset.Palette))
	copy(actualPalette, tileset.Palette)
	if tileset.AbsentTiles.Size() != 0 {
		actualPalette = addTransparent(tileset.Palette)
	}

	img := image.NewPaletted(image.Rect(0, 0, width*common.MetatileSizePx, height*common.MetatileSizePx),
		[]color.Color(actualPalette))

	x, y := 0, 0

	for _, mtile := range tileset.Metatiles {
		m.writeMetatileTile(tileset, img, mtile.TopLeft, actualPalette, x, y)
		m.writeMetatileTile(tileset, img, mtile.TopRight, actualPalette, x+common.TileSizePx, y)
		m.writeMetatileTile(tileset, img, mtile.BottomLeft, actualPalette, x, y+common.TileSizePx)
		m.writeMetatileTile(tileset, img, mtile.BottomRight, actualPalette, x+common.TileSizePx, y+common.TileSizePx)
		x += common.MetatileSizePx
		if x >= width*common.MetatileSizePx {
			x %= width * common.MetatileSizePx
			y += common.MetatileSizePx
		}
	}

	return img
}

func (m *Manager) writeMetatileTile(tileset *common.Metatiles, img *image.Paletted, index uint8, palette outPalette, x, y int) {
	refIt := tileset.Refs.Find(common.TileRef{Range: common.IndexRange{Start: index, End: index}})
	if refIt == nil {
		return
	}
	ref := refIt.GetValue()
	if len(ref.File) == 0 {
		return
	}

	tile, err := m.cache.getTile(ref.File, ref.Offset+(index-ref.Range.Start))
	if err != nil {
		return
	}

	writeTileToImage(img, palette, tile, x, y)

}

func (m *Manager) getOutPath(name, extension string, isTileData bool) string {
	isJSON := extension == common.ExtensionJSON
	return path.Join(m.out.GetOutputPath(isTileData, isJSON), name+extension)
}

type outPalette []color.Color

func addTransparent(palette []color.Color) outPalette {
	actualPalette := make(outPalette, 0, len(palette)+1)
	actualPalette = append(actualPalette, color.Transparent)
	actualPalette = append(actualPalette, palette...)
	return actualPalette
}

func (p outPalette) getColorIndex(rawIndex uint8) uint8 {
	if p[0] != color.Transparent {
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
