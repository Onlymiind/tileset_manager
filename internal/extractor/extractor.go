package extractor

import (
	"sort"

	"github.com/Onlymiind/tileset_manager/internal/common"
)

// Convert two bytes of source data to an array of color indexes
// lsb: byte with least significant bits of color index
// msb: byte with most significant bits of color index
// returns: slice of 8 color indexes
func getColorIndexes(lsb byte, msb byte) []byte {
	result := make([]byte, 8)
	mask := byte(1) << 7

	for i := 0; i < 8; i++ {
		result[i] = lsb&mask>>7 | msb&mask>>6

		lsb <<= 1
		msb <<= 1
	}

	return result
}

func getTile(src []byte) []byte {
	if len(src) != common.BytesPerTile {
		return nil
	}

	result := make([]byte, 0, common.BitsPerTile)
	for y := 0; y < common.TileSizePx; y++ {
		result = append(result, getColorIndexes(src[y*2], src[y*2+1])...)
	}

	return result
}

func ExtractTileData(src []byte) common.Tiles {
	tileCount := len(src) / common.BytesPerTile

	result := make(common.Tiles, 0, len(src)/common.BytesPerTile)

	for tile := 0; tile < tileCount; tile++ {
		offset := tile * common.BytesPerTile
		tileData := getTile(src[offset : offset+common.BytesPerTile])
		tileCopy := make([]byte, len(tileData))
		copy(tileCopy, tileData)
		result = append(result, tileCopy)
	}

	return result
}

func ExtractMetatileData(src []byte, tileData common.Tree[common.TileRef]) *common.Metatiles {
	if len(src) < 4 || len(src)%4 != 0 {
		return nil
	}

	result := common.NewMetatiles()

	result.Refs = tileData

	absent := map[uint8]struct{}{}
	for i := 0; i < len(src); i += 4 {
		tl, tr, bl, br := src[i], src[i+1], src[i+2], src[i+3]
		arr := []uint8{tl, tr, bl, br}
		for _, index := range arr {
			ref := common.TileRef{Range: common.IndexRange{Start: index, End: index}}
			it := tileData.Find(ref)
			if it == nil {
				absent[index] = struct{}{}
			} else if result.Refs.Find(ref) == nil {
				result.Refs.Insert(it.GetValue())
			}
		}

		result.Metatiles = append(result.Metatiles, common.Metatile{
			TopLeft:     tl,
			TopRight:    tr,
			BottomLeft:  bl,
			BottomRight: br,
		})
	}

	absentArr := make([]uint8, 0, len(absent))
	for i := range absent {
		absentArr = append(absentArr, i)
	}
	sort.Slice(absentArr, func(i, j int) bool { return absentArr[i] < absentArr[j] })

	absentRngs := make([]common.IndexRange, 0, len(absentArr))
	for i := range absentArr {
		if i == 0 || absentRngs[len(absentRngs)-1].End != absentArr[i]-1 {
			absentRngs = append(absentRngs, common.IndexRange{Start: absentArr[i], End: absentArr[i]})
		} else {
			absentRngs[len(absentRngs)-1].End = absentArr[i]
		}
	}

	for _, rng := range absentRngs {
		result.AbsentTiles.Insert(rng)
	}

	return result
}
