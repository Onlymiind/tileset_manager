package common

import (
	"image/color"
	"math"
	"path"
	"strings"

	"errors"
)

const (
	DefaultOutDir         = "extracted"
	DeafultCacheSizeKB    = 30
	ExtensionTileData     = ".chr"
	ExtensionMetatileData = ".mtile"
	ExtensionJSON         = ".json"
	ExtensionPNG          = ".png"
	OutTilesPerRow        = 16
	TileSizePx            = 8
	BitsPerTile           = TileSizePx * TileSizePx
	BytesPerTile          = TileSizePx * 2
	MetatileSizePx        = TileSizePx * 2

	ColorBlack     uint16 = 0
	ColorWhite     uint16 = 0xffff
	ColorDarkGray  uint16 = ColorWhite / 4
	ColorLightGray uint16 = ColorDarkGray * 2
)

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
	Type          OutputType
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
	Palette      []color.Color
	CacheSize    MemorySize
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

func (r *TileRef) InRange(index uint8) bool {
	return index >= r.Range.Start && index <= r.Range.End
}

type Tiles struct {
	Data    [][]byte
	Palette []color.Color
	Size    MemorySize
}

type Metatiles struct {
	Palette     []color.Color
	Refs        Tree[TileRef]
	AbsentTiles Tree[IndexRange]
	Metatiles   []Metatile
}

func NewMetatiles() *Metatiles {
	return &Metatiles{
		Refs:        NewTree(func(lhs, rhs *TileRef) bool { return lhs.Less(rhs) }),
		AbsentTiles: NewTree(func(lhs, rhs *IndexRange) bool { return lhs.Start < rhs.Start && lhs.End < rhs.End }),
	}
}

type MemoryUnit uint8

const (
	Bytes MemoryUnit = iota * 10
	Kilobytes
	Magebytes
	Gigabytes
	Terabytes
)

type MemorySize uint64

func MemorySizeFrom(value float64, unit MemoryUnit) MemorySize {
	return MemorySize(math.Round(value * float64(uint64(1)<<unit)))
}

func (s MemorySize) As(unit MemoryUnit) float64 {
	return float64(s) / float64(uint64(1)<<unit)
}

func (s MemorySize) Bytes() uint64 {
	return uint64(s)
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
