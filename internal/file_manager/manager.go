package file_manager

import (
	"errors"
	"fmt"
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

func getOutputFileName(srcPath string, newExtension string) string {
	if strings.Contains(srcPath, common.ExtensionTileData) {
		return strings.ReplaceAll(srcPath, common.ExtensionTileData, newExtension)
	} else if strings.Contains(srcPath, common.ExtensionMetatileData) {
		return strings.ReplaceAll(srcPath, common.ExtensionMetatileData, "_tileset"+newExtension)
	}

	return srcPath
}

func getOutputFilePath(outPath string, rootPath string, srcPath string, newExtension string) string {
	dest := strings.Replace(srcPath, rootPath, outPath, 1)
	return getOutputFileName(dest, newExtension)
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

func ReadFile(path string) ([]byte, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("error getting file info: %s", err.Error())
	}

	data := make([]byte, info.Size())
	read, err := file.Read(data)
	file.Close()
	if err != nil {
		return nil, fmt.Errorf("error reading file: %s", err.Error())
	} else if read != int(info.Size()) {
		return nil, errors.New("could not read all data")
	}

	return data, nil
}

func ExtractTileData(filePath string, palette [4]color.Color) (*proto.Tiles, error) {

	data, err := ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	pb := extractor.ExtractTileData(data)

	return pb, nil
}

func ExtractMetatileData(filePath string, tileData *proto.Tiles, palette [4]color.Color) (*proto.Tileset, error) {

	data, err := ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	pb := extractor.ExtractMetatileData(data, tileData, common.AirTileID, common.AirTileData[:])

	return pb, nil
}
