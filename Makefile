BINARY_NAME ?= studentbox
ENTRYPOINT ?= ./cmd/$(BINARY_NAME)/main.go
# tags come from: https://github.com/containers/podman/issues/12548#issuecomment-989053364
LIB_TAGS = remote exclude_graphdriver_btrfs btrfs_noversion exclude_graphdriver_devicemapper containers_image_openpgp
VERSION ?= $(shell git describe --tags --always --dirty)

build:
	@echo "Building $(BINARY_NAME)..."
	go build -tags "$(LIB_TAGS)" -ldflags '-X main.version=$(VERSION)' -o ./bin/$(BINARY_NAME) -v $(ENTRYPOINT)

download:
	@echo "Downloading dependencies..."
	go mod download