package file_manager

import (
	"errors"
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"os"
	"path"
	"strings"

	protobuf "github.com/golang/protobuf/proto"

	"github.com/Onlymiind/tileset_generator/internal/constants"
	"github.com/Onlymiind/tileset_generator/internal/extractor"
	"github.com/Onlymiind/tileset_generator/internal/image_writer"
	"github.com/Onlymiind/tileset_generator/proto"
	"github.com/golang/protobuf/jsonpb"
)

func IsTileData(info fs.FileInfo) bool {
	return strings.HasSuffix(info.Name(), constants.ExtensionTileData) && info.Mode().IsRegular()
}

func IsMetatileData(info fs.FileInfo) bool {
	return strings.HasSuffix(info.Name(), constants.ExtensionMetatileData) && info.Mode().IsRegular()
}

func writePB(srcFilePath string, msg protobuf.Message) error {
	pbPath := path.Join(constants.DefaultOutDir, strings.ReplaceAll(srcFilePath, constants.ExtensionTileData, ".json"))
	filePB, err := os.OpenFile(pbPath, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	marshaller := jsonpb.Marshaler{}
	err = marshaller.Marshal(filePB, msg)
	if err != nil {
		return err
	}
	return filePB.Close()
}

func writePNG(srcFilePath string, img *image.Paletted) error {
	imgPath := path.Join(constants.DefaultOutDir, strings.ReplaceAll(srcFilePath, constants.ExtensionTileData, ".png"))

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

func readFile(path string, info fs.FileInfo) ([]byte, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	data := make([]byte, info.Size())
	read, err := file.Read(data)
	file.Close()
	if err != nil {
		return nil, err
	} else if read != int(info.Size()) {
		return nil, errors.New("could not read all data")
	}

	return data, nil
}

func ExtractTileData(filePath string, info fs.FileInfo, palette [4]color.Color, outputPNG bool) (*proto.Tiles, error) {
	if !IsTileData(info) {
		return nil, nil
	}

	data, err := readFile(filePath, info)
	if err != nil {
		return nil, err
	}

	pb := extractor.ExtractTileData(data)

	if outputPNG {
		img := image_writer.WriteTileData(pb, constants.DefaultPalette)
		err = writePNG(filePath, img)
		if err != nil {
			return nil, err
		}
	}

	return pb, writePB(filePath, pb)
}
