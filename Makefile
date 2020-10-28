# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
CLI_BINARY_NAME=mixer-cli
API_BINARY_NAME=mixer-api

all: clean deps build
build-cli: deps
		$(GOBUILD) -o bin/$(CLI_BINARY_NAME) -v cmd/mixer-cli/main.go
build-api: deps
		$(GOBUILD) -o bin/$(API_BINARY_NAME) -v cmd/mixer-api/main.go
test:
		$(GOTEST) -v ./...
clean:
		$(GOCLEAN)
		rm -f $(CLI_BINARY_NAME)
		rm -f $(API_BINARY_NAME)
deps:
		$(GOGET) -u github.com/google/uuid

.PHONY: all build test clean deps
