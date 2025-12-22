.PHONY: all web client server clean build-dir

# Output directory for binaries
BUILD_DIR := bin

all: web client server

build-dir:
	mkdir -p $(BUILD_DIR)

web:
	cd web && npm install && npm run build

client: build-dir web
	go build -o $(BUILD_DIR)/client ./client

server: build-dir
	go build -o $(BUILD_DIR)/server ./server

clean:
	rm -rf $(BUILD_DIR)
	rm -rf web/dist
	rm -rf web/node_modules
