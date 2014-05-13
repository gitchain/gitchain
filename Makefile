SOURCES=$(wildcard *.go **/*.go)

all: gitchain

gitchain: $(SOURCES)
	go build

test:
	go test ./block ./transaction ./db
