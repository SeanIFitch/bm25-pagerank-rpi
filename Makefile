# Variables
BIN_DIR := ./bin
CMD_API := ./cmd/api/main.go
CMD_TRAIN := ./cmd/training/main.go
API_BIN := $(BIN_DIR)/api
TRAIN_BIN := $(BIN_DIR)/train
TEST_DIR := ./test

# Default target
.PHONY: all
all: build

# Run all tests in the project
.PHONY: test
test:
	@echo "Running all tests..."
	go test -v -cover ./...

# Build the binaries for the API and training commands
.PHONY: build
build:
	@echo "Building the API and training binaries..."
	mkdir -p $(BIN_DIR)
	go build -o $(API_BIN) $(CMD_API)
	go build -o $(TRAIN_BIN) $(CMD_TRAIN)

# Clean build artifacts (binaries, etc.)
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR)
