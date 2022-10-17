package file_manager

import (
	"errors"
	"image"
	"image/png"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"github.com/valyala/fastjson"
	"google.golang.org/protobuf/encoding/protojson"
	protobuf "google.golang.org/protobuf/proto"

	"github.com/Onlymiind/tileset_manager/internal/common"
	"github.com/Onlymiind/tileset_manager/internal/extractor"
	"github.com/Onlymiind/tileset_manager/proto"
)

type TileCache map[string][][]byte

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

		c[file] = tiles.Tiles
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

func ExtractTileData(filePath string) (*proto.Tiles, error) {

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	pb := extractor.ExtractTileData(data)

	return pb, nil
}

func ExtractMetatileData(filePath string, tileData *proto.Tiles, emptyTileID uint8, emptyTileData []byte) (*proto.Tileset, error) {

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	pb := extractor.ExtractMetatileData(data, tileData, emptyTileID, emptyTileData)

	return pb, nil
}

func getOutputType(t string) common.OutputType {
	switch t {
	case "png_only":
		return common.IgnoreJSON
	case "json_only":
		return common.IgnorePNG
	default:
		return 0
	}
}

func ParseConfig(cfgPath string) (*common.Config, error) {
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, common.Wrap(err, "could not read config file")
	}

	cfgJSON, err := fastjson.ParseBytes(data)
	if err != nil {
		return nil, common.Wrap(err, "could not parse config")
	}

	cfg := &common.Config{}
	cfg.Auto = string(cfgJSON.GetStringBytes("auto"))

	output := cfgJSON.GetObject("output")
	cfg.Output = common.Output{
		Directory:     string(output.Get("directory").GetStringBytes()),
		OutputType:    getOutputType(string(output.Get("type").GetStringBytes())),
		ImgDirectory:  string(output.Get("img_directory").GetStringBytes()),
		JSONDirectory: string(output.Get("json_directory").GetStringBytes()),
		TileDirectory: string(output.Get("tile_directory").GetStringBytes()),
	}

	output.Visit(func(idStr []byte, val *fastjson.Value) {
		id, err := strconv.ParseUint(string(idStr), 16, 8)
		if err == nil {
			cfg.EmptyTile.ID = uint8(id)
			cfg.EmptyTile.Data = string(val.GetStringBytes())
		}
	})

	convert := cfgJSON.GetArray("convert_to_png")
	cfg.ConvertToPng = make([]string, 0, len(convert))
	for i := range convert {
		cfg.ConvertToPng = append(cfg.ConvertToPng, string(convert[i].GetStringBytes()))
	}

	manual := cfgJSON.GetArray("manual")
	cfg.Manual = make([]common.Manual, 0, len(manual))
	for i := range manual {
		cfg.Manual = append(cfg.Manual, common.Manual{
			TileData:     string(manual[i].GetStringBytes("tile_data")),
			MetatileData: string(manual[i].GetStringBytes("metatile_data")),
			Name:         string(manual[i].GetStringBytes("name")),
		})
	}

	return cfg, nil
}
