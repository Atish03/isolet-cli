-include config.mk

# Default prefix
PREFIX ?= /usr/local
BIN_DIR = $(PREFIX)/bin
BUILD_FLAGS := -ldflags "-s -w"

BIN_NAME = isolet

all: build

build:
	go build $(BUILD_FLAGS) -o $(BIN_NAME) .

install: build
	mkdir -p $(BIN_DIR)
	cp $(BIN_NAME) $(BIN_DIR)/
	rm -f $(BIN_NAME)

uninstall:
	rm -f $(BIN_DIR)/$(BIN_NAME)

.PHONY: all build install uninstall clean
