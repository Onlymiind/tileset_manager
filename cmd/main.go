package main

import (
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/Onlymiind/tileset_generator/internal/constants"
	"github.com/Onlymiind/tileset_generator/internal/file_manager"
)

func FileWalker(filePath string, info fs.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if info.IsDir() {
		return os.MkdirAll(path.Join(constants.DefaultOutDir, filePath), 0777)
	}

	if file_manager.IsTileData(info) {
		_, err = file_manager.ExtractTileData(filePath, info, constants.DefaultPalette, true)
		return err
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("expected at least one argument")
	}

	err := os.Mkdir(constants.DefaultOutDir, 0777)
	if err != nil && !os.IsExist(err) {
		log.Fatal("can't create output directory")
	}

	err = filepath.Walk(os.Args[1], FileWalker)
	if err != nil {
		log.Fatal(err.Error())
	}
}
