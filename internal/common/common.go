package common

import (
	"image/color"
	"path"
	"strconv"
	"strings"

	"errors"
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

var EmptyIndexRange IndexRange = IndexRange{-1, -1}

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
	Start, End int
}

type TileRef struct {
	File     string
	Range    IndexRange
	RefRange IndexRange
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
	ids := strings.Split(indexes, ":")
	if len(ids) == 0 {
		return EmptyIndexRange, nil
	}
	start, err := strconv.ParseUint(ids[0], 16, 8)
	if err != nil {
		return EmptyIndexRange, Wrap(err, "could not convert index to integer")
	}
	end := -1
	if len(ids) > 1 {
		endId, err := strconv.ParseUint(ids[1], 16, 8)
		if err != nil {
			return EmptyIndexRange, Wrap(err, "could not convert index to integer")
		}
		end = int(endId)
	}
	return IndexRange{Start: int(start), End: end}, nil
}

func ParseTileRef(tileRange, refStr string) (*TileRef, error) {
	if len(tileRange) == 0 {
		return nil, errors.New("empty tile range")
	}
	indexes, err := ParseIndexRange(tileRange)
	if err != nil {
		return nil, Wrap(err, "could not parse tile range")
	} else if indexes == EmptyIndexRange {
		return nil, errors.New("empty tile range")
	}

	path, refRange, _ := strings.Cut(refStr, ":")
	if len(path) == 0 {
		return nil, errors.New("empty file path")
	}

	ref, err := ParseIndexRange(refRange)
	if err != nil {
		return nil, errors.New("could not parse tile range")
	}

	return &TileRef{
		File:     path,
		Range:    indexes,
		RefRange: ref,
	}, nil
}
