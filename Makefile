# Makefile

# Variables
BIN_DIR := ./bin
CMD_API := ./cmd/api/main.go
CMD_TRAIN := ./cmd/train/main.go
API_BIN := $(BIN_DIR)/api
TRAIN_BIN := $(BIN_DIR)/train

# Default target (optional)
.PHONY: all
all: build

# Run all tests in the project
.PHONY: test
test:
	@echo "Running all tests..."
	go test ./...

# Build the binaries for the API and training commands
.PHONY: build
build:
	@echo "Building the API and training binaries..."
	mkdir -p $(BIN_DIR)
	go build -o $(API_BIN) $(CMD_API)
	go build -o $(TRAIN_BIN) $(CMD_TRAIN)

# Start the API server
.PHONY: run-api
run-api: build
	@echo "Starting the API server..."
	$(API_BIN)

# Train the ranking model
.PHONY: train
train: build
	@echo "Training the ranking model..."
	$(TRAIN_BIN)

# Clean build artifacts (binaries, etc.)
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR)

# Generate raw data for Microsoft and RPI
.PHONY: generate-data
generate-data:
	@echo "Generating raw data..."
	./scripts/generate_microsoft_data.sh
	./scripts/generate_query_data.sh

# Generate processed data for queries
.PHONY: generate-processed-data
generate-processed-data:
	@echo "Generating processed query data..."
	./scripts/train_model.sh
