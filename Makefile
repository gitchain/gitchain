SOURCES=$(wildcard *.go **/*.go)

all: gitchain

gitchain: $(SOURCES) ui/bindata.go
	@go build

test:
	@go test ./keys ./router ./block ./transaction ./db

ui/bindata.go: ui $(filter-out ui/bindata.go, $(wildcard ui/**)) Makefile
	@go-bindata -pkg=ui -o=ui/bindata.go -ignore=\(bindata.go\|\.gitignore\) -prefix=ui ui
