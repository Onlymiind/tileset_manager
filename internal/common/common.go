package common

import (
	"encoding/base64"
	"image/color"
	"path"
	"strconv"
	"strings"

	"errors"

	"github.com/valyala/fastjson"
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

type OutputType uint8

const (
	IgnorePNG OutputType = 1 << iota
	IgnoreJSON
)

type Output struct {
	Directory     string
	ImgDirectory  string
	JSONDirectory string
	TileDirectory string
	OutputType    OutputType
}

func (o *Output) GetOutputPath(isTile bool, isJSON bool) string {
	out := []string{
		o.Directory,
	}
	if isTile && len(o.TileDirectory) != 0 {
		out = append(out, o.TileDirectory)
	}

	if isJSON && len(o.JSONDirectory) != 0 {
		out = append(out, o.JSONDirectory)
	} else if !isJSON && len(o.ImgDirectory) != 0 {
		out = append(out, o.ImgDirectory)
	}

	return path.Join(out...)
}

type Config struct {
	Auto         string
	Output       Output
	Manual       []Manual
	ConvertToPng []string
	EmptyTile    EmptyTile
}

type EmptyTile struct {
	ID   uint8
	Data string
}

type Manual struct {
	TileData     string
	MetatileData string
	Name         string
}

type IndexRange struct {
	Start, End uint8
}

type TileRef struct {
	File   string
	Range  IndexRange
	Offset uint8
}

func (r TileRef) Less(rhs TileRef) bool {
	return r.Range.Start < rhs.Range.Start && r.Range.End < rhs.Range.End
}

type Metatiles struct {
	Palette [4]color.Color
}

func ReplaceLast(src string, old string, new string) string {
	i := strings.LastIndex(src, old)
	if i < 0 {
		return src
	}
	return src[:i] + strings.ReplaceAll(src[i:], old, new)
}

func Wrap(err error, msgs ...string) error {
	if err != nil {
		msgs = append(msgs, err.Error())
	}
	return errors.New(strings.Join(msgs, ": "))
}

func ParseIndexRange(indexes string) (IndexRange, error) {
	first, last, found := strings.Cut(indexes, ":")
	if len(first) == 0 {
		return IndexRange{}, errors.New("invalid range")
	}
	start, err := strconv.ParseUint(first, 16, 8)
	if err != nil {
		return IndexRange{}, Wrap(err, "could not convert index to integer")
	}
	if len(last) == 0 {
		end := uint8(start)
		if found {
			end = ^uint8(0)
		}
		return IndexRange{Start: uint8(start), End: end}, nil
	}

	end, err := strconv.ParseUint(first, 16, 8)
	if err != nil {
		return IndexRange{}, Wrap(err, "could not convert index to integer")
	}
	return IndexRange{Start: uint8(start), End: uint8(end)}, nil
}

func ParseTileRef(tileRange, refStr string) (*TileRef, error) {
	if len(tileRange) == 0 {
		return nil, errors.New("empty tile range")
	}
	indexes, err := ParseIndexRange(tileRange)
	if err != nil {
		return nil, Wrap(err, "could not parse tile range")
	}

	path, offsetStr, _ := strings.Cut(refStr, ":")
	if len(path) == 0 {
		return nil, errors.New("empty file path")
	}

	offset := uint8(0)
	if len(offsetStr) != 0 {
		offsetU64, err := strconv.ParseUint(offsetStr, 16, 8)
		if err != nil {
			return nil, Wrap(err, "could not parse offset")
		}
		offset = uint8(offsetU64)
	}

	return &TileRef{
		File:   path,
		Range:  indexes,
		Offset: offset,
	}, nil
}

func SerializeTileData(data [][]byte) *fastjson.Value {
	arena := fastjson.Arena{}
	result := arena.NewArray()

	for i, tile := range data {
		result.SetArrayItem(i, arena.NewString(base64.StdEncoding.EncodeToString(tile)))
	}

	return result
}
