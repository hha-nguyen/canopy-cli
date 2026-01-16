.PHONY: build build-all install test clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

BINARY_NAME := canopy
OUTPUT_DIR := bin

build:
	@mkdir -p $(OUTPUT_DIR)
	go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME) .

build-all:
	@mkdir -p $(OUTPUT_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "Built binaries in $(OUTPUT_DIR)"
	@ls -la $(OUTPUT_DIR)/$(BINARY_NAME)-*

install: build
	cp $(OUTPUT_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to $(GOPATH)/bin"

test:
	go test -v ./...

clean:
	rm -rf $(OUTPUT_DIR)

release: build-all checksums tarballs

tarballs:
	@cd $(OUTPUT_DIR) && \
	for f in $(BINARY_NAME)-darwin-* $(BINARY_NAME)-linux-*; do \
		tar -czf "$${f}.tar.gz" "$$f" && rm "$$f"; \
	done
	@cd $(OUTPUT_DIR) && \
	if [ -f "$(BINARY_NAME)-windows-amd64.exe" ]; then \
		zip -q "$(BINARY_NAME)-windows-amd64.zip" "$(BINARY_NAME)-windows-amd64.exe" && \
		rm "$(BINARY_NAME)-windows-amd64.exe"; \
	fi
	@echo "Created release archives"

checksums:
	@cd $(OUTPUT_DIR) && \
	rm -f checksums.txt && \
	for f in $(BINARY_NAME)-*; do \
		shasum -a 256 "$$f" >> checksums.txt; \
	done
	@echo "Checksums written to $(OUTPUT_DIR)/checksums.txt"
