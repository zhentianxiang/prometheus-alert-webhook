# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=prometheus-webhook
BINARY_DIR=bin

# Target platforms (GOOS/GOARCH)
PLATFORMS := darwin/amd64 linux/amd64

.PHONY: all build clean run test

all: build

build:
	@echo "Building for all platforms..."
	$(foreach platform, $(PLATFORMS), $(call build_platform, $(platform)))

# A helper function to build for a specific platform
define build_platform
	$(eval parts := $(subst /, ,$(1)))
	$(eval GOOS := $(word 1, $(parts)))
	$(eval GOARCH := $(word 2, $(parts)))
	@echo "Building for $(GOOS)/$(GOARCH)..."
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -o $(BINARY_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH) main.go
endef

run:
	@echo "Running the application..."
	$(GOCMD) run main.go

clean:
	@echo "Cleaning up..."
	$(GOCLEAN)
	rm -rf $(BINARY_DIR)

test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

default: build
