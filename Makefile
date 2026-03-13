.PHONY: setup proto build build-server embed-web web dev-tui dev-web dev-server doc-dev doc-build test test-unit test-int lint fmt clean install

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X github.com/KashifKhn/kassie/internal/shared/version.Version=$(VERSION) \
           -X github.com/KashifKhn/kassie/internal/shared/version.Commit=$(COMMIT) \
           -X github.com/KashifKhn/kassie/internal/shared/version.BuildDate=$(BUILD_DATE)

setup:
	@echo "Installing protoc plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Setup complete!"

proto:
	@echo "Generating protobuf code..."
	./scripts/gen-proto.sh

build: lint test web embed-web
	@echo "Building kassie binary with embedded web assets..."
	go build -ldflags="$(LDFLAGS)" -o kassie cmd/kassie/main.go
	@echo "Cleaning up copied assets..."
	@find internal/server/web/dist -mindepth 1 ! -name '.gitkeep' -delete 2>/dev/null || true

embed-web:
	@echo "Copying web assets for embedding..."
	@find internal/server/web/dist -mindepth 1 ! -name '.gitkeep' -delete 2>/dev/null || true
	@cp -r web/dist/* internal/server/web/dist/ 2>/dev/null || true

build-server:
	@echo "Building server only..."
	go build -ldflags="$(LDFLAGS)" -tags=noui -o kassie cmd/kassie/main.go

web:
	@echo "Building web UI..."
	@if [ -d "web" ]; then \
		cd web && pnpm install && pnpm run build; \
	else \
		echo "Web directory not found, skipping"; \
	fi

dev-tui:
	@echo "Running TUI in development..."
	go run cmd/kassie/main.go tui

dev-web:
	@echo "Running web UI on port 9091..."
	@echo "Make sure to run 'make dev-server' in another terminal"
	cd web && pnpm dev

dev-server:
	@echo "Running server for web development on port 9090..."
	go run cmd/kassie/main.go server --http-port 9090

doc-dev:
	@echo "Running docs dev server..."
	cd docs && pnpm dev

doc-build:
	@echo "Building docs..."
	cd docs && pnpm install && pnpm run build

test:
	@echo "Running all tests..."
	go test ./...

test-unit:
	@echo "Running unit tests..."
	go test -short ./...

test-int:
	@echo "Running integration tests..."
	go test -tags=integration ./...

lint:
	@echo "Running linters..."
	golangci-lint run

fmt:
	@echo "Formatting code..."
	go fmt ./...
	@if [ -f web/package.json ]; then cd web && pnpm run format; fi

clean:
	@echo "Cleaning build artifacts..."
	rm -f kassie
	rm -rf web/dist
	rm -rf api/gen/go/*
	rm -rf api/gen/ts/*

install: build
	@echo "Installing kassie to /usr/local/bin..."
	@cp kassie /usr/local/bin/
	@echo "Installation complete!"
