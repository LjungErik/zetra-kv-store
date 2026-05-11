BINARY_NAME := kv-node
IMAGE_NAME  := zetra-kv-store
IMAGE_TAG   := latest
BIN_DIR     := bin

.PHONY: all build docker-build up down logs clean

all: build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/kv-node

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) -f cmd/kv-node/Dockerfile .

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

clean:
	rm -rf $(BIN_DIR)
