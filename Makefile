SOURCES=$(wildcard *.go **/*.go)

all: gitchain

gitchain: $(SOURCES)
	go build

test:
	go test ./keys ./block ./transaction ./db
