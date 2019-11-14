all: build

build:
	GO111MODULE=on go build -o dist/cgw cmd/cmd.go
