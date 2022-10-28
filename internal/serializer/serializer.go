package serializer

import (
	"encoding/base64"
	"fmt"
	"image/color"

	"github.com/Onlymiind/tileset_manager/internal/common"
	"github.com/valyala/fastjson"
)

func SerializeTileData(data *common.Tiles) *fastjson.Value {
	arena := &fastjson.Arena{}
	result := arena.NewObject()
	result.Set(fileType, arena.NewString(typeTileData))

	arr := arena.NewArray()
	for i, tile := range data.Data {
		arr.SetArrayItem(i, arena.NewString(base64.StdEncoding.EncodeToString(tile)))
	}
	result.Set(tiles, arr)
	plt := arena.NewArray()
	for i, color := range data.Palette {
		plt.SetArrayItem(i, serializeColor(data.Palette, arena, color))
	}
	result.Set(palette, plt)

	_ = color.RGBA{}

	return result
}

func SerializeMetatileData(plt []color.Color, data *common.Metatiles) *fastjson.Value {
	arena := &fastjson.Arena{}
	result := arena.NewObject()
	result.Set(fileType, arena.NewString(typeMetatileData))

	absTiles := arena.NewArray()
	i := 0
	for it := data.AbsentTiles.Begin(); it != nil; it = it.Next() {
		absTiles.SetArrayItem(i, arena.NewString(serializeTileRange(it.GetValue())))
		i++
	}
	result.Set(absentTiles, absTiles)

	tileRefs := arena.NewObject()
	for it := data.Refs.Begin(); it != nil; it = it.Next() {
		tileRefs.Set(serializeTileRef(arena, it.GetValue()))
	}
	result.Set(tiles, tileRefs)

	metatiles := arena.NewArray()
	for i := range data.Metatiles {
		metatiles.SetArrayItem(i, serializeMetatile(arena, data.Metatiles[i]))
	}
	result.Set(mtiles, metatiles)

	if len(data.Palette) != 0 {
		paletteObj := arena.NewArray()
		for i := range data.Palette {
			paletteObj.SetArrayItem(i, serializeColor(plt, arena, data.Palette[i]))
		}
		result.Set(palette, paletteObj)
	}

	return result
}

func serializeMetatile(arena *fastjson.Arena, mtile common.Metatile) *fastjson.Value {
	result := arena.NewObject()
	result.Set(topLeft, arena.NewString(fmt.Sprintf("%x", mtile.TopLeft)))
	result.Set(topRight, arena.NewString(fmt.Sprintf("%x", mtile.TopRight)))
	result.Set(bottomLeft, arena.NewString(fmt.Sprintf("%x", mtile.BottomLeft)))
	result.Set(bottomRight, arena.NewString(fmt.Sprintf("%x", mtile.BottomRight)))

	return result
}

func serializeTileRange(rng common.IndexRange) string {
	if rng.Start == rng.End {
		return fmt.Sprintf("%x", rng.Start)
	} else {
		return fmt.Sprintf("%x:%x", rng.Start, rng.End)
	}
}

func serializeTileRef(arena *fastjson.Arena, ref common.TileRef) (key string, refStr *fastjson.Value) {
	key = serializeTileRange(ref.Range)

	if ref.Offset != 0 {
		refStr = arena.NewString(fmt.Sprintf("%s:%x", ref.File, ref.Offset))
	} else {
		refStr = arena.NewString(ref.File)
	}

	return key, refStr
}

func serializeColor(palette []color.Color, arena *fastjson.Arena, c color.Color) *fastjson.Value {
	model := color.Palette(palette)
	r, g, b, _ := model.Convert(c).RGBA()

	return arena.NewString(fmt.Sprintf("%02x%02x%02x", uint8(r), uint8(g), uint8(b)))
}
