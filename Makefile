BINARY := ollamon
BUILD_DIR := bin
DIST_DIR := dist
MAIN_PKG := ./cmd/ollamon
GOCACHE := $(CURDIR)/.gocache
VERSION := 0.1.0

.PHONY: build run tidy clean dist

build:
	mkdir -p $(BUILD_DIR)
	mkdir -p $(GOCACHE)
	GOCACHE=$(GOCACHE) GOPROXY=off go build -o $(BUILD_DIR)/$(BINARY) $(MAIN_PKG)

run:
	mkdir -p $(GOCACHE)
	GOCACHE=$(GOCACHE) GOPROXY=off go run $(MAIN_PKG)

tidy:
	mkdir -p $(GOCACHE)
	GOCACHE=$(GOCACHE) go mod tidy

clean:
	rm -rf $(BUILD_DIR) $(GOCACHE)

dist:
	mkdir -p $(GOCACHE)
	$(MAKE) tidy
	mkdir -p $(DIST_DIR)
	@for target in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64; do \
		os=$${target%/*}; \
		arch=$${target#*/}; \
		echo "Building $$target..."; \
		tmpdir=$$(mktemp -d); \
		GOCACHE=$(GOCACHE) GOPROXY=off GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=v$(VERSION)" -o $$tmpdir/$(BINARY) $(MAIN_PKG); \
		tar -C $$tmpdir -czf $(DIST_DIR)/$(BINARY)_$(VERSION)_$$os\_$$arch.tar.gz $(BINARY); \
		rm -rf $$tmpdir; \
	done
	shasum -a 256 $(DIST_DIR)/*.tar.gz > $(DIST_DIR)/checksums.txt
	@echo "Release artifacts in $(DIST_DIR)/"
	@ls -lh $(DIST_DIR)/
