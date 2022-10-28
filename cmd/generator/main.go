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
	"github.com/Onlymiind/tileset_manager/internal/serializer"
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

	cfg, err := serializer.ParseConfig(os.Args[1])
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

	manager := file_manager.NewManager(cfg)
	fileWalkerWrapper := func(filePath string, info fs.FileInfo, err error) error {
		return fileWalker(cfg, manager, filePath, info, err)
	}

	if len(cfg.Auto) != 0 {
		err = filepath.Walk(cfg.Auto, fileWalkerWrapper)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	fmt.Printf("%f %s\n", manager.CacheSize().As(common.Kilobytes), "kb")

	processManual(cfg, manager)

	fmt.Printf("%f %s\n", manager.CacheSize().As(common.Kilobytes), "kb")

	processConvertToPNG(cfg, manager)

	fmt.Printf("%f %s\n", manager.CacheSize().As(common.Kilobytes), "kb")

	// f, _ := os.OpenFile("out/png/queen.png", os.O_RDONLY, 0666)
	// img, _ := png.Decode(f)

	// _ = img

	// fmt.Println()
}

func process(cfg *common.Config, manager *file_manager.Manager, tilePath, metatilePath, name string, writeTileData bool) error {
	tileData, err := file_manager.ExtractTileData(tilePath)
	if err != nil {
		return common.Wrap(err, "failed to extract tile data", tilePath)
	}
	tileData.Palette = cfg.Palette

	if writeTileData {
		json := serializer.SerializeTileData(tileData)
		err = manager.WriteJSON(json, name+".tile", true)
		if err != nil {
			return common.Wrap(err, "failed to write json", tilePath)
		}

		png := file_manager.TileDataToImage(tileData)
		err = manager.WritePNG(png, name, true)
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
				End:   uint8(len(tileData.Data)),
			},
		})
		if len(cfg.EmptyTile.File) != 0 {
			refs.Insert(cfg.EmptyTile)
		}

		mtiles, err := file_manager.ExtractMetatileData(metatilePath, refs)
		if err != nil {
			return common.Wrap(err, "failed to extract metatile data")
		}
		if len(mtiles.Palette) == 0 {
			mtiles.Palette = cfg.Palette
		}

		json := serializer.SerializeMetatileData(cfg.Palette, mtiles)
		err = manager.WriteJSON(json, name+".mtile", false)
		if err != nil {
			return common.Wrap(err, "failed to write json", tilePath)
		}

		png := manager.MetatileToImage(mtiles)
		err = manager.WritePNG(png, name, false)
		if err != nil {
			return common.Wrap(err, "failed to write png", tilePath)
		}
	}

	return nil
}

func processManual(cfg *common.Config, manager *file_manager.Manager) {
	for i := range cfg.Manual {
		info, err := os.Stat(cfg.Manual[i].TileData)
		if err != nil {
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

		err = process(cfg, manager, cfg.Manual[i].TileData, metatilePath, name, false)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func processConvertToPNG(cfg *common.Config, manager *file_manager.Manager) {
	for i := range cfg.ConvertToPng {
		info, err := os.Stat(cfg.ConvertToPng[i])
		if err != nil {
			fmt.Printf("failed to get file info %s, error: %s\n", cfg.ConvertToPng[i], err.Error())
		}
		name := info.Name()
		for ext := path.Ext(name); len(ext) != 0; ext = path.Ext(name) {
			name = strings.TrimSuffix(name, ext)
		}
		if strings.HasSuffix(cfg.ConvertToPng[i], ".tile.json") {
			tileData, err := serializer.ParseTileData(cfg.ConvertToPng[i])
			if err != nil {
				fmt.Println(err.Error(), cfg.ConvertToPng[i])
			}

			if len(tileData.Palette) == 0 {
				tileData.Palette = cfg.Palette
			}

			img := file_manager.TileDataToImage(tileData)
			err = manager.WritePNG(img, name, true)
			if err != nil {
				fmt.Println(err.Error(), cfg.ConvertToPng[i])
			}

		} else if strings.HasSuffix(cfg.ConvertToPng[i], ".mtile.json") {
			tileset, err := serializer.ParseMetatileData(cfg.ConvertToPng[i])
			if err != nil {
				fmt.Println(err.Error(), cfg.ConvertToPng[i])
			}

			if len(tileset.Palette) == 0 {
				tileset.Palette = cfg.Palette
			}

			img := manager.MetatileToImage(tileset)
			err = manager.WritePNG(img, name, false)
			if err != nil {
				fmt.Println(err.Error(), cfg.ConvertToPng[i])
			}
		}
	}
}

func fileWalker(cfg *common.Config, manager *file_manager.Manager, filePath string, info fs.FileInfo, err error) error {
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
		return process(cfg, manager, filePath, metatilePath, name, true)
	}

	return nil
}
