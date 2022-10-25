package serializer

import (
	"errors"
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"

	"github.com/Onlymiind/tileset_manager/internal/common"
	"github.com/valyala/fastjson"
)

func ParseIndexRange(indexes string) (common.IndexRange, error) {
	first, last, found := strings.Cut(indexes, ":")
	if len(first) == 0 {
		return common.IndexRange{}, errors.New("invalid range")
	}
	start, err := strconv.ParseUint(first, 16, 8)
	if err != nil {
		return common.IndexRange{}, common.Wrap(err, "could not convert index to integer")
	}
	if len(last) == 0 {
		end := uint8(start)
		if found {
			end = ^uint8(0)
		}
		return common.IndexRange{Start: uint8(start), End: end}, nil
	}

	end, err := strconv.ParseUint(first, 16, 8)
	if err != nil {
		return common.IndexRange{}, common.Wrap(err, "could not convert index to integer")
	}
	return common.IndexRange{Start: uint8(start), End: uint8(end)}, nil
}

func ParseTileRef(tileRange, refStr string) (*common.TileRef, error) {
	if len(tileRange) == 0 {
		return nil, errors.New("empty tile range")
	}
	indexes, err := ParseIndexRange(tileRange)
	if err != nil {
		return nil, common.Wrap(err, "could not parse tile range")
	}

	path, offsetStr, _ := strings.Cut(refStr, ":")
	if len(path) == 0 {
		return nil, errors.New("empty file path")
	}

	offset := uint8(0)
	if len(offsetStr) != 0 {
		offsetU64, err := strconv.ParseUint(offsetStr, 16, 8)
		if err != nil {
			return nil, common.Wrap(err, "could not parse offset")
		}
		offset = uint8(offsetU64)
	}

	return &common.TileRef{
		File:   path,
		Range:  indexes,
		Offset: offset,
	}, nil
}

func ParseConfig(cfgPath string) (*common.Config, error) {
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, common.Wrap(err, "could not read config file")
	}

	cfgJSON, err := fastjson.ParseBytes(data)
	if err != nil {
		return nil, common.Wrap(err, "could not parse config")
	}

	cfg := &common.Config{}
	cfg.Auto = string(cfgJSON.GetStringBytes("auto"))

	output := cfgJSON.GetObject("output")
	cfg.Output = common.Output{
		Directory:     string(output.Get("directory").GetStringBytes()),
		OutputType:    getOutputType(string(output.Get("type").GetStringBytes())),
		ImgDirectory:  string(output.Get("img_directory").GetStringBytes()),
		JSONDirectory: string(output.Get("json_directory").GetStringBytes()),
		TileDirectory: string(output.Get("tile_directory").GetStringBytes()),
	}

	cfgJSON.GetObject("empty_tile").Visit(func(idStr []byte, val *fastjson.Value) {
		ref, err := ParseTileRef(string(idStr), string(val.GetStringBytes()))
		if err == nil {
			cfg.EmptyTile = *ref
		}
	})

	convert := cfgJSON.GetArray("convert_to_png")
	cfg.ConvertToPng = make([]string, 0, len(convert))
	for i := range convert {
		cfg.ConvertToPng = append(cfg.ConvertToPng, string(convert[i].GetStringBytes()))
	}

	manual := cfgJSON.GetArray("manual")
	cfg.Manual = make([]common.Manual, 0, len(manual))
	for i := range manual {
		cfg.Manual = append(cfg.Manual, common.Manual{
			TileData:     string(manual[i].GetStringBytes("tile_data")),
			MetatileData: string(manual[i].GetStringBytes("metatile_data")),
			Name:         string(manual[i].GetStringBytes("name")),
		})
	}

	return cfg, nil
}

func ParseMetatileData(path string) (*common.Metatiles, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, common.Wrap(err, "could not read file", path)
	}

	less := func(lhs, rhs *common.TileRef) bool { return lhs.Less(rhs) }
	result := &common.Metatiles{
		Refs:        common.NewTree(less),
		AbsentTiles: common.NewTree(less),
	}

	parsed, err := fastjson.ParseBytes(data)
	if err != nil {
		return nil, common.Wrap(err, "could not parse metatile data", path)
	}

	palette := parsed.GetArray("palette")
	if palette != nil && len(palette) != 4 {
		return nil, fmt.Errorf("wrong palette length: %d, expected 4: %s", len(palette), path)
	}

	for i := range palette {
		result.Palette[i] = parseColor(string(palette[i].GetStringBytes()))
	}

	parsed.GetObject("tiles").Visit(func(ids []byte, refStr *fastjson.Value) {
		ref, err := ParseTileRef(string(ids), string(refStr.GetStringBytes()))
		if err == nil {
			result.Refs.Insert(*ref)
		}
	})
	parsed.GetObject("absent_tiles").Visit(func(ids []byte, refStr *fastjson.Value) {
		ref, err := ParseTileRef(string(ids), string(refStr.GetStringBytes()))
		if err == nil {
			result.AbsentTiles.Insert(*ref)
		}
	})

	metatiles := parsed.GetArray("metatiles")
	result.Metatiles = make([]common.Metatile, 0, len(metatiles))
	for i := range metatiles {
		mtile := common.Metatile{
			TopLeft:     uint8(metatiles[i].GetInt("tl")),
			TopRight:    uint8(metatiles[i].GetInt("tr")),
			BottomLeft:  uint8(metatiles[i].GetInt("bl")),
			BottomRight: uint8(metatiles[i].GetInt("br")),
		}
		result.Metatiles = append(result.Metatiles, mtile)
	}

	return result, nil
}

func getOutputType(t string) common.OutputType {
	switch t {
	case "png_only":
		return common.IgnoreJSON
	case "json_only":
		return common.IgnorePNG
	default:
		return 0
	}
}

func parseColor(str string) color.Color {
	switch str {
	case "black", "00":
		return color.Gray16{common.ColorBlack}
	case "white", "11":
		return color.Gray16{common.ColorWhite}
	case "light", "10":
		return color.Gray16{common.ColorLightGray}
	case "dark", "01":
		return color.Gray16{common.ColorDarkGray}
	default:
		return color.Gray16{common.ColorBlack}
	}
}
