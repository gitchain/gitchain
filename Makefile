SOURCES=$(wildcard **/*.go)

all: gitchain

gitchain: $(SOURCES)
	go build

test:
	go test
