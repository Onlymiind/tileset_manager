package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	protobuf "github.com/golang/protobuf/proto"

	"github.com/Onlymiind/tileset_generator/internal/common"
	"github.com/Onlymiind/tileset_generator/internal/file_manager"
	"github.com/Onlymiind/tileset_generator/internal/image_writer"
	"github.com/Onlymiind/tileset_generator/proto"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("expected path to a config file as an argument")
	}

	defer func() {
		err := recover()
		if err != nil {
			log.Println(os.Getwd())
			log.Println(os.Args)
			log.Fatalln(err)
		}
	}()

	cfg, err := getConfig(os.Args[1])
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = os.MkdirAll(cfg.OutputDirectory, 0777)
	if err != nil {
		log.Fatal("could not create output directory")
	}

	fileWalkerWrapper := func(filePath string, info fs.FileInfo, err error) error {
		return fileWalker(cfg.OutputDirectory, cfg.Auto, "", cfg.IgnoreMetatiles, cfg.OutputType, filePath, info, err)
	}

	if len(cfg.Auto) != 0 {
		err = filepath.Walk(cfg.Auto, fileWalkerWrapper)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	processManual(cfg)

	processConvertToPNG(cfg)

}

type OutputType uint8

const (
	IgnorePNG OutputType = 1 << iota
	IgnoreJSON
)

type Config struct {
	Auto            string     `json:"auto,omitempty"`
	IgnoreMetatiles bool       `json:"ignore_metatiles,omitempty"`
	Manual          []Manual   `json:"manual,omitempty"`
	OutputDirectory string     `json:"output_directory,omitempty"`
	OutputType      OutputType `json:"output_type,omitempty"`
	ConvertToPng    []string   `json:"convert_to_png,omitempty"`
}

type Manual struct {
	TileData     string `json:"tile_data,omitempty"`
	MetatileData string `json:"metatile_data,omitempty"`
	Name         string `json:"name,omitempty"`
}

func getConfig(path string) (*Config, error) {
	cfgData, err := os.ReadFile(os.Args[1])
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %s", err.Error())
	}

	cfg := &Config{}
	err = json.Unmarshal(cfgData, &cfg)
	if err != nil {
		return nil, fmt.Errorf("could not parse config file: %s", err.Error())
	}

	if len(cfg.OutputDirectory) == 0 {
		cfg.OutputDirectory = common.DefaultOutDir
	}

	cfg.OutputDirectory = strings.TrimSuffix(cfg.OutputDirectory, "/")
	if len(cfg.OutputDirectory) == 0 {
		cfg.OutputDirectory = "."
	}

	cfg.Auto = strings.TrimSuffix(cfg.Auto, "/")

	if len(cfg.Auto) == 0 && len(cfg.Manual) == 0 {
		return nil, errors.New("no files to process")
	}

	return cfg, nil
}

func process(destFile, tileDataPath, metatileDataPath string, outputType OutputType) error {
	pngOut := destFile + ".png"
	jsonOut := destFile + common.ExtensionJSON

	tileData, err := file_manager.ExtractTileData(tileDataPath, common.DefaultPalette)
	if err != nil {
		return err
	}

	var imgOut *image.Paletted
	if outputType&IgnorePNG == 0 {
		imgOut = image_writer.WriteTileData(tileData, common.DefaultPalette)
	}
	var pbOut protobuf.Message = tileData

	if len(metatileDataPath) != 0 {
		tileset, err := file_manager.ExtractMetatileData(metatileDataPath, tileData, common.DefaultPalette)
		if err != nil {
			return err
		}

		pbOut = tileset

		if outputType&IgnorePNG == 0 {
			imgOut = image_writer.WriteMetatileData(tileset, common.DefaultPalette)
		}
	}

	if outputType&IgnoreJSON == 0 {
		err = file_manager.WritePB(jsonOut, pbOut)
		if err != nil {
			return err
		}
	}
	if outputType&IgnorePNG == 0 {
		err = file_manager.WritePNG(pngOut, imgOut)
		if err != nil {
			return err
		}
	}

	return nil
}

func processManual(cfg *Config) {
	for i := range cfg.Manual {
		info, err := os.Stat(cfg.Manual[i].TileData)
		if err != nil || !file_manager.IsTileData(info) {
			fmt.Printf("could not get tile data file info, path: %s, error: %s\n", cfg.Manual[i].TileData, err.Error())
			continue
		}

		name := strings.TrimSuffix(info.Name(), common.ExtensionTileData)

		metatilePath := ""
		if cfg.Manual[i].MetatileData != "" {
			info, err := os.Stat(cfg.Manual[i].MetatileData)
			if err != nil {
				fmt.Printf("could not get metatile data file info, path: %s, error %s\n", cfg.Manual[i].MetatileData, err.Error())
			} else if file_manager.IsMetatileData(info) {
				metatilePath = cfg.Manual[i].MetatileData
				name = strings.TrimSuffix(info.Name(), common.ExtensionMetatileData)
			}
		}

		if len(cfg.Manual[i].Name) != 0 {
			name = cfg.Manual[i].Name
		}

		dest := path.Join(cfg.OutputDirectory, name)

		err = process(dest, cfg.Manual[i].TileData, metatilePath, cfg.OutputType)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func processConvertToPNG(cfg *Config) {
	for i := range cfg.ConvertToPng {
		data, err := os.ReadFile(cfg.ConvertToPng[i])
		if err != nil {
			fmt.Printf("could not read file: %s\n", err.Error())
			continue
		}

		var tileset *proto.Tileset
		reader := bytes.NewReader(data)

		tileData := &proto.Tiles{}
		err = jsonpb.Unmarshal(reader, tileData)
		if err != nil {
			tileData = nil
			reader.Seek(0, io.SeekStart)
			tileset = &proto.Tileset{}
			err = jsonpb.Unmarshal(reader, tileset)
			if err != nil {
				fmt.Printf("could not unmarshal json: %s\n", err.Error())
				continue
			}
		}

		outFile := cfg.ConvertToPng[i]
		info, err := os.Stat(cfg.ConvertToPng[i])
		if err == nil {
			outFile = path.Join(cfg.OutputDirectory, info.Name())
		}
		outFile = common.ReplaceLast(outFile, common.ExtensionJSON, ".png")

		if tileData != nil {
			img := image_writer.WriteTileData(tileData, common.DefaultPalette)
			err = file_manager.WritePNG(outFile, img)
			if err != nil {
				fmt.Printf("could not write file: %s\n", err.Error())
			}
		}

		if tileset != nil {
			img := image_writer.WriteMetatileData(tileset, common.DefaultPalette)
			err = file_manager.WritePNG(outFile, img)
			if err != nil {
				fmt.Printf("could not write file: %s\n", err.Error())
			}
		}
	}
}

func getOutFilePath(outPath, rootPath, srcPath, oldName, newName, extensionToStrip string) string {
	dest := strings.Replace(srcPath, rootPath, outPath, 1)
	if len(newName) != 0 {
		dest = common.ReplaceLast(dest, oldName, newName)
	} else {
		dest = strings.TrimSuffix(dest, extensionToStrip)
	}

	return dest
}

func fileWalker(outPath, rootPath, newName string, ingoreMetatiles bool, outputType OutputType, filePath string, info fs.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if filePath == rootPath {
		return nil
	}

	if info.IsDir() {
		return os.MkdirAll(strings.Replace(filePath, rootPath, outPath, 1), 0777)
	}

	if file_manager.IsTileData(info) {

		dest := getOutFilePath(outPath, rootPath, filePath, info.Name(), newName, common.ExtensionTileData)

		metatilePath := common.ReplaceLast(filePath, common.ExtensionTileData, common.ExtensionMetatileData)
		if info, err := os.Stat(metatilePath); ingoreMetatiles || !(err == nil && file_manager.IsMetatileData(info)) {
			metatilePath = ""
		}

		return process(dest, filePath, metatilePath, outputType)
	}

	return nil
}