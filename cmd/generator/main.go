package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Onlymiind/tileset_manager/internal/common"
	"github.com/Onlymiind/tileset_manager/internal/file_manager"
	"github.com/Onlymiind/tileset_manager/internal/image_writer"
	"github.com/Onlymiind/tileset_manager/internal/serializer"
)

func main() {
	// if len(os.Args) < 2 {
	// 	log.Fatalln("expected path to a config file as an argument")
	// }

	defer func() {
		err := recover()
		if err != nil {
			log.Println(os.Getwd())
			log.Println(os.Args)
			log.Fatalln(err)
		}
	}()

	cfg, err := serializer.ParseConfig("assets/config.cfg.json")
	if err != nil {
		log.Fatalln(err.Error())
	}

	outDirs := []string{
		cfg.Output.GetOutputPath(false, false),
		cfg.Output.GetOutputPath(true, false),
		cfg.Output.GetOutputPath(false, true),
		cfg.Output.GetOutputPath(true, true),
	}

	for i := range outDirs {
		if len(outDirs[i]) == 0 {
			continue
		}
		err = os.MkdirAll(outDirs[i], 0777)
		if err != nil {
			log.Fatalln("could not create output directory", outDirs[i])
		}
	}

	cache := file_manager.NewTileCache()
	fileWalkerWrapper := func(filePath string, info fs.FileInfo, err error) error {
		return fileWalker(cfg, &cache, filePath, info, err)
	}

	if len(cfg.Auto) != 0 {
		err = filepath.Walk(cfg.Auto, fileWalkerWrapper)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	processManual(cfg, &cache)

	processConvertToPNG(cfg, &cache)

	// f, _ := os.OpenFile("out/png/queen.png", os.O_RDONLY, 0666)
	// img, _ := png.Decode(f)

	// _ = img

	// fmt.Println()
}

func process(cfg *common.Config, cache *file_manager.TileCache, tilePath, metatilePath, name string, writeTileData bool) error {
	tileData, err := file_manager.ExtractTileData(tilePath)
	if err != nil {
		return common.Wrap(err, "failed to extract tile data", tilePath)
	}

	jsonPath := path.Join(cfg.Output.GetOutputPath(true, true), name+common.ExtensionJSON)

	if writeTileData {
		json := serializer.SerializeTileData(tileData)
		err = serializer.WriteJson(jsonPath, json)
		if err != nil {
			return common.Wrap(err, "failed to write json", tilePath)
		}

		png := image_writer.WriteTileData(tileData, common.DefaultPalette)
		err = serializer.WritePng(path.Join(cfg.Output.GetOutputPath(true, false), name+common.ExtensionPNG), png)
		if err != nil {
			return common.Wrap(err, "failed to write png", tilePath)
		}
	}
	if len(metatilePath) != 0 {
		refs := common.NewTree(func(lhs, rhs *common.TileRef) bool { return lhs.Less(rhs) })
		refs.Insert(common.TileRef{
			File: tilePath,
			Range: common.IndexRange{
				Start: 0,
				End:   uint8(len(tileData)),
			},
		})
		if len(cfg.EmptyTile.File) != 0 {
			refs.Insert(cfg.EmptyTile)
		}

		mtiles, err := file_manager.ExtractMetatileData(metatilePath, refs)
		if err != nil {
			return common.Wrap(err, "failed to extract metatile data")
		}

		json := serializer.SerializeMetatileData(mtiles)
		err = serializer.WriteJson(path.Join(cfg.Output.GetOutputPath(false, true), name+common.ExtensionJSON), json)
		if err != nil {
			return common.Wrap(err, "failed to write json", metatilePath)
		}

		if len(mtiles.Palette) == 0 {
			mtiles.Palette = common.DefaultPalette[:]
		}

		png := image_writer.WriteMetatileData(cache, mtiles)
		err = serializer.WritePng(path.Join(cfg.Output.GetOutputPath(false, false), name+common.ExtensionPNG), png)
		if err != nil {
			return common.Wrap(err, "failed to write png", metatilePath)
		}
	}

	return nil
}

func processManual(cfg *common.Config, cache *file_manager.TileCache) {
	for i := range cfg.Manual {
		info, err := os.Stat(cfg.Manual[i].TileData)
		if err != nil || !file_manager.IsTileData(info) {
			fmt.Printf("could not get tile data file info, path: %s, error: %s\n", cfg.Manual[i].TileData, err.Error())
			continue
		}
		name := strings.TrimSuffix(info.Name(), path.Ext(info.Name()))

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

		mInfo, err := os.Stat(metatilePath)
		if err != nil || !file_manager.IsMetatileData(mInfo) {
			metatilePath = ""
		}

		if len(cfg.Manual[i].Name) != 0 {
			name = cfg.Manual[i].Name
		}

		err = process(cfg, cache, cfg.Manual[i].TileData, metatilePath, name, true)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func processConvertToPNG(cfg *common.Config, cache *file_manager.TileCache) {
	for i := range cfg.ConvertToPng {
		info, err := os.Stat(cfg.ConvertToPng[i])
		if err != nil {
			fmt.Printf("failed to get file info %s, error: %s\n", cfg.ConvertToPng[i], err.Error())
		}
		name := strings.TrimSuffix(info.Name(), path.Ext(info.Name())) + common.ExtensionPNG
		if strings.HasSuffix(cfg.ConvertToPng[i], ".tile.json") {
			tileData, err := serializer.ParseTileData(cfg.ConvertToPng[i])
			if err != nil {
				fmt.Println(err.Error(), cfg.ConvertToPng[i])
			}

			img := image_writer.WriteTileData(tileData, common.DefaultPalette)
			err = serializer.WritePng(path.Join(cfg.Output.GetOutputPath(true, false), name), img)
			if err != nil {
				fmt.Println(err.Error(), cfg.ConvertToPng[i])
			}

		} else if strings.HasSuffix(cfg.ConvertToPng[i], ".mtile.json") {
			tileset, err := serializer.ParseMetatileData(cfg.ConvertToPng[i])
			if err != nil {
				fmt.Println(err.Error(), cfg.ConvertToPng[i])
			}

			if len(tileset.Palette) == 0 {
				tileset.Palette = common.DefaultPalette[:]
			}

			img := image_writer.WriteMetatileData(cache, tileset)
			err = serializer.WritePng(path.Join(cfg.Output.GetOutputPath(false, false), name), img)
			if err != nil {
				fmt.Println(err.Error(), cfg.ConvertToPng[i])
			}
		}
	}
}

func fileWalker(cfg *common.Config, cache *file_manager.TileCache, filePath string, info fs.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if file_manager.IsTileData(info) {

		name := strings.TrimSuffix(info.Name(), path.Ext(info.Name()))
		metatilePath := common.ReplaceLast(filePath, common.ExtensionTileData, common.ExtensionMetatileData)
		mInfo, err := os.Stat(metatilePath)
		if err != nil || !file_manager.IsMetatileData(mInfo) {
			metatilePath = ""
		}
		return process(cfg, cache, filePath, metatilePath, name, true)
	}

	return nil
}
