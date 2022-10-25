package serializer

import (
	"encoding/base64"

	"github.com/valyala/fastjson"
)

func SerializeTileData(data [][]byte) *fastjson.Value {
	arena := fastjson.Arena{}
	result := arena.NewArray()

	for i, tile := range data {
		result.SetArrayItem(i, arena.NewString(base64.StdEncoding.EncodeToString(tile)))
	}

	return result
}
