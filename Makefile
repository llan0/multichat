.PHONY: build run run-pprof stats

BINARY_NAME=multichat
BINARY_PATH=./cmd/

build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) $(BINARY_PATH)
	@echo "Build complete: ./$(BINARY_NAME)"

run:
	@echo "Running $(BINARY_NAME)..."
	@go run $(BINARY_PATH)

run-pprof:
	@echo "Running $(BINARY_NAME) with pprof enabled..."
	@echo "Profiling: http://localhost:6060/debug/pprof/"
	@go run $(BINARY_PATH) -pprof

stats:
	@echo "Checking process stats..."
	@PID=$$(pgrep -f multichat | head -1); \
	if [ -z "$$PID" ]; then \
		echo "multichat process not found"; \
	else \
		echo "PID: $$PID"; \
		ps -p $$PID -o pid,%cpu,%mem,rss,vsz,time; \
	fi
