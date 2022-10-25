package common

import (
	"image/color"
	"path"
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

	ColorBlack     uint16 = 0
	ColorWhite     uint16 = 0xffff
	ColorDarkGray  uint16 = ColorWhite / 4
	ColorLightGray uint16 = ColorDarkGray * 2
)

var AirTileData [BitsPerTile]byte = [BitsPerTile]byte{}
var DefaultPalette [4]color.Color = [4]color.Color{
	color.Gray16{ColorBlack},
	color.Gray16{ColorWhite},
	color.Gray16{ColorLightGray}, //Light gray
	color.Gray16{ColorDarkGray},  //Dark gray
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
	EmptyTile    TileRef
}

type Manual struct {
	TileData     string
	MetatileData string
	Name         string
}

type IndexRange struct {
	Start, End uint8
}

type Metatile struct {
	TopLeft     uint8
	TopRight    uint8
	BottomLeft  uint8
	BottomRight uint8
}

type TileRef struct {
	File   string
	Range  IndexRange
	Offset uint8
}

func (r *TileRef) Less(rhs *TileRef) bool {
	return r.Range.Start < rhs.Range.Start && r.Range.End < rhs.Range.End
}

type Tiles [][]byte

type Metatiles struct {
	Palette     [4]color.Color
	Refs        Tree[TileRef]
	AbsentTiles Tree[TileRef]
	Metatiles   []Metatile
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
