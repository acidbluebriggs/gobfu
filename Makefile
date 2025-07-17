.PHONY: build clean run install

GOBFU=gobfu
BUILD_DIR=./bin

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(GOBFU) ./cmd/gobfu


run: build
	$(BUILD_DIR)/$(GOBFU)
	$(MAKE) clean

clean:
	rm -rf $(BUILD_DIR)

# Build, run, and clean in one command
build-run-clean: build run clean

# Install to GOPATH/bin
install:
	go install ./cmd/gobfu