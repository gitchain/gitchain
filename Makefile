SOURCES=$(wildcard *.go **/*.go)

all: gitchain

gitchain: $(SOURCES)
	@go build

test:
	@go test ./keys ./router ./block ./transaction ./db
