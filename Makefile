all: compiler generator

compiler:

generator:
	go build -o bin/tileset_manager cmd/generator/main.go