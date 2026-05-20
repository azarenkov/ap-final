.PHONY: proto build test up down logs lint

PROTOC_PATH := $(HOME)/go/bin

proto:
	@command -v protoc >/dev/null || { echo "install protoc: brew install protobuf"; exit 1; }
	@for f in train user booking notification; do \
		mkdir -p gen/$$f/v1; \
		PATH=$(PROTOC_PATH):$$PATH protoc -I=proto \
			--go_out=gen --go_opt=module=github.com/azarenkov/ap2-final-gen \
			--go-grpc_out=gen --go-grpc_opt=module=github.com/azarenkov/ap2-final-gen \
			proto/$$f/v1/$$f.proto; \
	done
	cd gen && go mod tidy

build:
	cd train-service && go build ./...
	cd api-gateway   && go build ./...

test:
	cd train-service && go test -race -count=1 ./...
	cd api-gateway   && go test -race -count=1 ./...

up:
	docker compose up -d --build

down:
	docker compose down -v

logs:
	docker compose logs -f train-service api-gateway

lint:
	cd train-service && go vet ./...
	cd api-gateway   && go vet ./...
