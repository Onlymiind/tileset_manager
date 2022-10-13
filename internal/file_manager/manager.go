package file_manager

import (
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"os"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	protobuf "google.golang.org/protobuf/proto"

	"github.com/Onlymiind/tileset_generator/internal/common"
	"github.com/Onlymiind/tileset_generator/internal/extractor"
	"github.com/Onlymiind/tileset_generator/proto"
)

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

func ExtractTileData(filePath string, palette [4]color.Color) (*proto.Tiles, error) {

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	pb := extractor.ExtractTileData(data)

	return pb, nil
}

func ExtractMetatileData(filePath string, tileData *proto.Tiles, emptyTileID uint8, emptyTileData []byte, palette [4]color.Color) (*proto.Tileset, error) {

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	pb := extractor.ExtractMetatileData(data, tileData, emptyTileID, emptyTileData)

	return pb, nil
}
