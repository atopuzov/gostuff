GOFILES := $(wildcard *.go)
TARGET := sq_measure

VERSION := 1.0.0
BUILD := $(shell git rev-parse HEAD)

LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"

all: $(TARGET);

$(TARGET): $(GOFILES)
	@go build $(LDFLAGS) -o $(TARGET)

.PHONY: clean
clean:
	@rm -f $(TARGET)


