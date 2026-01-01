BINARY_NAME=multichat
BINARY_PATH=./cmd/

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build -o $(BINARY_NAME) $(BINARY_PATH)

.PHONY: run
run:
	go run $(BINARY_PATH)

.PHONY: clean
clean:
	rm -f $(BINARY_NAME)
