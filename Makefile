all: proto _compiler _generator

compiler: proto _compiler

generator: proto _generator

proto:
	protoc --proto_path=proto --go_out=proto --go_opt=paths=source_relative proto/*.proto

_compiler:

_generator:
	go build -o bin/tileset_generator cmd/generator/main.go

.PHONY: proto