.PHONY: proto build

all: proto build

proto: pkg/wikigraphpb/wikigraph.proto
	@echo "Generating proto objects..."
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./pkg/wikigraphpb/wikigraph.proto

build:
	go build -o bin/server cmd/server/*
	go build -o bin/client cmd/client/*
	go build -o bin/worker cmd/worker/*

docker-publish-x86:
	docker buildx build --platform linux/x86_64 -t lodthe/wikigraph-server -f dockerfiles/Dockerfile-server .
	docker push lodthe/wikigraph-server
	@echo "Server image published"

	docker buildx build --platform linux/x86_64 -t lodthe/wikigraph-client -f dockerfiles/Dockerfile-client .
	docker push lodthe/wikigraph-client
	@echo "Client image published"

	docker buildx build --platform linux/x86_64 -t lodthe/wikigraph-worker -f dockerfiles/Dockerfile-worker .
	docker push lodthe/wikigraph-worker
	@echo "Worker image published"
