# Variables
BIN_DIR := ./bin
CMD_API := ./cmd/api/main.go
API_BIN := $(BIN_DIR)/api
GOARCH := amd64      # Target x86-64 architecture (Intel/AMD 64-bit)
GOOS := linux        # Target Linux OS

# Default target
.PHONY: all
all: build

# Run all tests in the project
.PHONY: test
test:
	@echo "Running all tests..."
	@go test -v -cover ./... | grep -v "\--- PASS:"

# Build the binaries for the API
.PHONY: build
build:
	@echo "Building the API binary..."
	mkdir -p $(BIN_DIR)
	GOARCH=$(GOARCH) GOOS=$(GOOS) go build -o $(API_BIN) $(CMD_API)

# Clean build artifacts (binaries, etc.)
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR)
