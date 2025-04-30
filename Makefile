# Makefile for MC PeePee

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTIDY=$(GOCMD) mod tidy
GOCLEAN=$(GOCMD) clean

# Binary name
BINARY_NAME=mcpeepee
BINARY_UNIX=$(BINARY_NAME)
# If you want to support Windows cross-compilation uncomment below
# GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME).exe main.go

all: tidy build

.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BINARY_UNIX) .
	@echo "Build complete: $(BINARY_UNIX)"

.PHONY: run
run: build
	@echo "Running $(BINARY_UNIX)..."
	./$(BINARY_UNIX)

.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	$(GOTIDY)

.PHONY: clean
clean:
	@echo "Cleaning up..."
	$(GOCLEAN)
	rm -f $(BINARY_UNIX)
	rm -f $(BINARY_NAME).exe # If supporting Windows
	@echo "Cleanup complete."

.PHONY: help
help:
	@echo "Available commands:"
	@echo "  all       - Tidy dependencies and build the binary (default)"
	@echo "  build     - Build the binary"
	@echo "  run       - Build and run the binary"
	@echo "  tidy      - Tidy Go module dependencies"
	@echo "  clean     - Remove built binary and clean cache" 