all: compiler generator

compiler:

generator:
	go build -o bin/tileset_manager_w -ldflags=-w cmd/generator/main.go