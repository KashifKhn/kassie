.PHONY: setup proto build build-server web dev-tui dev-web dev-server test test-unit test-int lint fmt clean

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

build: web
	@echo "Building kassie binary..."
	go build -o kassie cmd/kassie/main.go

build-server:
	@echo "Building server only..."
	go build -tags=noui -o kassie cmd/kassie/main.go

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
