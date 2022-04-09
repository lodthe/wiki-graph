.PHONY: proto build

all: proto build

proto: pkg/wikigraphpb/wikigraph.proto
	@echo "Generating proto objects..."
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./pkg/wikigraphpb/wikigraph.proto

build:
	go build -o bin/server cmd/server/*
	go build -o bin/client cmd/client/*
