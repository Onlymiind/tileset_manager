package file_manager

import (
	"errors"
	"image"
	"image/png"
	"io/fs"
	"os"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	protobuf "google.golang.org/protobuf/proto"

	"github.com/Onlymiind/tileset_manager/internal/common"
	"github.com/Onlymiind/tileset_manager/internal/extractor"
	"github.com/Onlymiind/tileset_manager/proto"
)

type TileCache map[string]common.Tiles

func NewTileCache() TileCache {
	return make(TileCache)
}

func (c TileCache) GetTile(file string, index uint8) ([]byte, error) {
	data, ok := c[file]
	if !ok {
		tiles, err := ExtractTileData(file)
		if err != nil {
			return nil, common.Wrap(err, "cache", "could not get tile data")
		}

		c[file] = tiles
		data = c[file]
	}

	if int(index) >= len(data) {
		return nil, errors.New("tile index out of bounds")
	}

	return data[index], nil
}

func IsTileData(info fs.FileInfo) bool {
	return strings.HasSuffix(info.Name(), common.ExtensionTileData) && info.Mode().IsRegular()
}

func IsMetatileData(info fs.FileInfo) bool {
	return strings.HasSuffix(info.Name(), common.ExtensionMetatileData) && info.Mode().IsRegular()
}

func WritePB(pbPath string, msg protobuf.Message) error {
	marshaller := protojson.MarshalOptions{
		EmitUnpopulated: true,
		Multiline:       true,
		Indent:          "    ",
	}
	out, err := marshaller.Marshal(msg)
	if err != nil {
		return err
	}
	return os.WriteFile(pbPath, out, 0666)
}

func WritePNG(imgPath string, img *image.Paletted) error {

	imgFile, err := os.OpenFile(imgPath, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	err = png.Encode(imgFile, img)
	if err != nil {
		return err
	}
	return imgFile.Close()
}

func ExtractTileData(filePath string) (common.Tiles, error) {

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	tileData := extractor.ExtractTileData(data)

	return tileData, nil
}

func ExtractMetatileData(filePath string, tileData common.Tiles, emptyTileID uint8, emptyTileData []byte) (*proto.Tileset, error) {

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	pb := extractor.ExtractMetatileData(data, tileData, emptyTileID, emptyTileData)

	return pb, nil
}
