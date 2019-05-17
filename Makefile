all: format test build

format:
	go fmt ./...

test:
	go test ./...

build:
	go build ./cmd/clustertest
	go build ./cmd/clustertestd
