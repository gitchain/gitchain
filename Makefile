SOURCES=$(wildcard *.go **/*.go **/**/*.go)

all: gitchain

gitchain: $(SOURCES) ui/bindata.go
	@go build

test:
	@go test ./keys ./block ./transaction ./db ./git

ui/bindata.go: ui $(filter-out ui/bindata.go, $(wildcard ui/**)) Makefile
	@go-bindata -pkg=ui -o=ui/bindata.go -ignore=\(bindata.go\|\.gitignore\) -prefix=ui ui

prepare:
	@[ `pwd` = $(GOPATH)/src/github.com/gitchain/gitchain ] || (echo "Make sure you check out this repository to $(GOPATH)/src/github.com/gitchain/gitchain" && exit 1) 
	@mkdir -p $(GOPATH)/bin
	@go get github.com/jteeuwen/go-bindata/go-bindata
	@go get github.com/tools/godep
	@[ -f $(GOPATH)/bin/godep ] || (echo "godep failed to install" && exit 1)
	@[ -f $(GOPATH)/bin/go-bindata ] || (echo "go-bindata failed to install" && exit 1) 
	@godep restore
