package serializer

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/Onlymiind/tileset_manager/internal/common"
	"github.com/valyala/fastjson"
)

func SerializeTileData(data common.Tiles) *fastjson.Value {
	arena := fastjson.Arena{}
	result := arena.NewArray()

	for i, tile := range data {
		result.SetArrayItem(i, arena.NewString(base64.StdEncoding.EncodeToString(tile)))
	}

	return result
}

func SerializeMetatileData(data *common.Metatiles) *fastjson.Value {
	arena := &fastjson.Arena{}
	result := arena.NewObject()

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
			paletteObj.SetArrayItem(i, serializeColor(arena, data.Palette[i]))
		}
		result.Set(palette, paletteObj)
	}

	return result
}

func WritePng(path string, img *image.Paletted) error {
	buf := &bytes.Buffer{}
	err := png.Encode(buf, img)
	if err != nil {
		return common.Wrap(err, "failed to encode image", path)
	}
	err = os.WriteFile(path, buf.Bytes(), 0666)
	if err != nil {
		return common.Wrap(err, "failed to write to file", path)
	}

	return nil
}

func WriteJson(path string, json *fastjson.Value) error {
	err := os.WriteFile(path, json.MarshalTo(nil), 0666)
	if err != nil {
		return common.Wrap(err, "failed to write to file", path)
	}

	return nil
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

func serializeColor(arena *fastjson.Arena, c color.Color) *fastjson.Value {
	model := color.Palette(common.DefaultPalette[:])
	c16, ok := model.Convert(c).(color.Gray16)
	if !ok {
		return nil
	}

	colorStr := "black"
	switch c16.Y {
	case common.ColorBlack:
		colorStr = "black"
	case common.ColorWhite:
		colorStr = "white"
	case common.ColorLightGray:
		colorStr = "light"
	case common.ColorDarkGray:
		colorStr = "dark"
	}

	return arena.NewString(colorStr)

}
