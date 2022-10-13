package extractor

import (
	"github.com/Onlymiind/tileset_generator/internal/common"
	"github.com/Onlymiind/tileset_generator/proto"
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

func writeTileToProto(msg *proto.Tiles, tile []byte) {
	if len(tile) != common.BitsPerTile {
		return
	}

	tileCopy := make([]byte, len(tile))
	copy(tileCopy, tile)
	msg.Tiles = append(msg.Tiles, tileCopy)
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

func ExtractTileData(src []byte) *proto.Tiles {
	tileCount := len(src) / common.BytesPerTile

	encoded := &proto.Tiles{
		Tiles: make([][]byte, 0, len(src)/common.BytesPerTile),
	}

	for tile := 0; tile < tileCount; tile++ {
		offset := tile * common.BytesPerTile
		tileData := getTile(src[offset : offset+common.BytesPerTile])
		writeTileToProto(encoded, tileData)
	}

	return encoded
}

func ExtractMetatileData(src []byte, tileData *proto.Tiles, emptyTileID byte, emptyTileData []byte) *proto.Tileset {
	if len(src) < 4 || len(src)%4 != 0 {
		return nil
	}

	result := &proto.Tileset{
		TileData:  make(map[uint32][]byte, len(tileData.Tiles)+1),
		Metatiles: make([]*proto.Metatile, 0, len(src)/4),
	}

	for i := range tileData.Tiles {
		result.TileData[uint32(i)] = tileData.Tiles[i]
	}

	if len(emptyTileData) == common.BitsPerTile {
		result.TileData[uint32(emptyTileID)] = emptyTileData
	}

	for i := 0; i < len(src); i += 4 {
		result.Metatiles = append(result.Metatiles, &proto.Metatile{
			TopLeft:     uint32(src[i]),
			TopRight:    uint32(src[i+1]),
			BottomLeft:  uint32(src[i+2]),
			BottomRight: uint32(src[i+3]),
		})
	}

	return result
}
