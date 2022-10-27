package serializer

import (
	"encoding/base64"
	"errors"
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"

	"github.com/Onlymiind/tileset_manager/internal/common"
	"github.com/valyala/fastjson"
)

func ParseTileData(path string) (common.Tiles, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, common.Wrap(err, "could not read file", path)
	}

	json, err := fastjson.ParseBytes(data)
	if err != nil {
		return nil, common.Wrap(err, "could not parse json", path)
	}
	arr := json.GetArray()
	result := make(common.Tiles, 0, len(arr))
	for i := range arr {
		decoded, err := base64.StdEncoding.DecodeString(string(arr[i].GetStringBytes()))
		if err == nil {
			result = append(result, decoded)
		}
	}

	return result, nil
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
	cfg.Auto = string(cfgJSON.GetStringBytes(auto))

	output := cfgJSON.GetObject(out)
	cfg.Output = common.Output{
		Directory:     string(output.Get(outDir).GetStringBytes()),
		Type:          getOutputType(string(output.Get(outType).GetStringBytes())),
		ImgDirectory:  string(output.Get(imgDir).GetStringBytes()),
		JSONDirectory: string(output.Get(jsonDir).GetStringBytes()),
		TileDirectory: string(output.Get(tileDir).GetStringBytes()),
	}

	cfgJSON.GetObject(emptyTile).Visit(func(idStr []byte, val *fastjson.Value) {
		ref, err := parseTileRef(string(idStr), string(val.GetStringBytes()))
		if err == nil {
			cfg.EmptyTile = *ref
		}
	})

	convert := cfgJSON.GetArray(convertToPng)
	cfg.ConvertToPng = make([]string, 0, len(convert))
	for i := range convert {
		cfg.ConvertToPng = append(cfg.ConvertToPng, string(convert[i].GetStringBytes()))
	}

	manual := cfgJSON.GetArray(manual)
	cfg.Manual = make([]common.Manual, 0, len(manual))
	for i := range manual {
		cfg.Manual = append(cfg.Manual, common.Manual{
			TileData:     string(manual[i].GetStringBytes(tileData)),
			MetatileData: string(manual[i].GetStringBytes(mtileData)),
			Name:         string(manual[i].GetStringBytes(name)),
		})
	}

	return cfg, nil
}

func ParseMetatileData(path string) (*common.Metatiles, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, common.Wrap(err, "could not read file", path)
	}

	result := common.NewMetatiles()

	parsed, err := fastjson.ParseBytes(data)
	if err != nil {
		return nil, common.Wrap(err, "could not parse metatile data", path)
	}

	palette := parsed.GetArray(palette)
	if palette != nil && len(palette) != 4 {
		return nil, fmt.Errorf("wrong palette length: %d, expected 4: %s", len(palette), path)
	}

	for i := range palette {
		result.Palette[i] = parseColor(string(palette[i].GetStringBytes()))
	}

	parsed.GetObject(tiles).Visit(func(ids []byte, refStr *fastjson.Value) {
		ref, err := parseTileRef(string(ids), string(refStr.GetStringBytes()))
		if err == nil {
			result.Refs.Insert(*ref)
		}
	})
	absent := parsed.GetArray(absentTiles)
	for i := range absent {
		rng, err := parseIndexRange(string(absent[i].GetStringBytes()))
		if err == nil {
			result.AbsentTiles.Insert(rng)
		}
	}

	metatiles := parsed.GetArray(mtiles)
	result.Metatiles = make([]common.Metatile, 0, len(metatiles))
	for i := range metatiles {
		tl, err := strconv.ParseUint(string(metatiles[i].GetStringBytes(topLeft)), 16, 8)
		if err != nil {
			continue
		}
		tr, err := strconv.ParseUint(string(metatiles[i].GetStringBytes(topRight)), 16, 8)
		if err != nil {
			continue
		}
		bl, err := strconv.ParseUint(string(metatiles[i].GetStringBytes(bottomLeft)), 16, 8)
		if err != nil {
			continue
		}
		br, err := strconv.ParseUint(string(metatiles[i].GetStringBytes(bottomRight)), 16, 8)
		if err != nil {
			continue
		}
		mtile := common.Metatile{
			TopLeft:     uint8(tl),
			TopRight:    uint8(tr),
			BottomLeft:  uint8(bl),
			BottomRight: uint8(br),
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

func parseIndexRange(indexes string) (common.IndexRange, error) {
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

	end, err := strconv.ParseUint(last, 16, 8)
	if err != nil {
		return common.IndexRange{}, common.Wrap(err, "could not convert index to integer")
	}
	return common.IndexRange{Start: uint8(start), End: uint8(end)}, nil
}

func parseTileRef(tileRange, refStr string) (*common.TileRef, error) {
	if len(tileRange) == 0 {
		return nil, errors.New("empty tile range")
	}
	indexes, err := parseIndexRange(tileRange)
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
