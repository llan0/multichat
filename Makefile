.PHONY: build run

BINARY_NAME=multichat
BINARY_PATH=./cmd/multichat

build: 
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) $(BINARY_PATH)
	@echo "Build complete: ./$(BINARY_NAME)"

run: 
	@echo "Running $(BINARY_NAME)..."
	@go run $(BINARY_PATH)
