BINARY := ollamon
BUILD_DIR := bin
MAIN_PKG := ./cmd/ollamon

.PHONY: build run tidy clean

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) $(MAIN_PKG)

run:
	go run $(MAIN_PKG)

tidy:
	go mod tidy

clean:
	rm -rf $(BUILD_DIR)
