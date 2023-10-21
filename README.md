## About

This is a small command-line utility for converting graphics data in Game Boy's format to PNG. The program expects path to config JSON to be passed as a parameter.

## Configuring

- output.directory - base directory for the output
- output.img_directory - directory to output PNG files into
- output.tile_directory - base directory for decoded tiles
- output.json_directory - directory for JSON-encoded output (for metatiles includes indidies of not found tiles, see below)
- output.type - one of the "png_only", "json_only", "png_and_json"
- palette - array of four hex-encoded RGB colors.
- cache_size - controls the amout of memory used by loaded tile data when decoding metatiles 

The effective path for PNGs is <output.directory>/<output.img_directory> for metatiles and <output.directory>/<output.tile_directory>/<output.img_directory> for tiles.

- "auto" - contents for this directory will be automatically processed. That is, all files with the .chr extension are treated as tile data and all files with .mtile extension are treated as metatile data. The program tries to decode each .mtile file using .chr file with the same name. Any tile indicies that are missing from .chr file are written to "absent" array in resulting JSON and corresponding metatile is omitted from PNG.
- "manual" - use to manually map .chr file to .mtile file, as well as assign custom name to the outputted files. Check schemas/config.json for format.
- "convert_to_png" - list of files with JSON-encoded metatile data to convert to PNG image. Check schemas/metatiles.json for format.
