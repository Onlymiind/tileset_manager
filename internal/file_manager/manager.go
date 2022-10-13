package file_manager

import (
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"os"
	"strings"

	protobuf "github.com/golang/protobuf/proto"

	"github.com/Onlymiind/tileset_generator/internal/common"
	"github.com/Onlymiind/tileset_generator/internal/extractor"
	"github.com/Onlymiind/tileset_generator/proto"
	"github.com/golang/protobuf/jsonpb"
)

func IsTileData(info fs.FileInfo) bool {
	return strings.HasSuffix(info.Name(), common.ExtensionTileData) && info.Mode().IsRegular()
}

func IsMetatileData(info fs.FileInfo) bool {
	return strings.HasSuffix(info.Name(), common.ExtensionMetatileData) && info.Mode().IsRegular()
}

func WritePB(pbPath string, msg protobuf.Message) error {
	filePB, err := os.OpenFile(pbPath, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	marshaller := jsonpb.Marshaler{
		EmitDefaults: true,
		Indent:       "    ",
	}
	err = marshaller.Marshal(filePB, msg)
	if err != nil {
		return err
	}
	return filePB.Close()
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

func ExtractMetatileData(filePath string, tileData *proto.Tiles, palette [4]color.Color) (*proto.Tileset, error) {

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	pb := extractor.ExtractMetatileData(data, tileData, common.AirTileID, common.AirTileData[:])

	return pb, nil
}
