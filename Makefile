all: build test

test:
	dist/cgw --config example/config.yaml

build:
	GO111MODULE=on go build -o dist/cgw cmd/cmd.go
