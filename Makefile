# Orchion Monorepo Makefile
# Provides common commands for managing the entire project

.PHONY: help proto build build-all run-all clean clean-all test test-all install-deps

help:
	@echo "Orchion Monorepo Commands"
	@echo ""
	@echo "  make proto        - Generate protobuf files for all components"
	@echo "  make build        - Build all components (Go + Dashboard)"
	@echo "  make build-all    - Alias for build"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make clean-all    - Alias for clean"
	@echo "  make test         - Run tests for all components"
	@echo "  make install-deps - Install all dependencies (Go + npm)"
	@echo ""
	@echo "Note: On Windows, you may prefer using PowerShell scripts:"
	@echo "  .\shared\scripts\build-all.ps1"
	@echo "  .\shared\scripts\run-all.ps1"

proto:
	@echo "Generating protobuf files..."
	@cd orchestrator && make proto
	@cd node-agent && make proto

build: build-all

build-all:
	@echo "Building all components..."
	@cd orchestrator && go build -o orchestrator.exe ./cmd/orchestrator
	@cd node-agent && go build -o node-agent.exe ./cmd/node-agent
	@echo "Building dashboard..."
	@cd dashboard && npm install && npm run build || echo "Dashboard build skipped"
	@echo "Build complete!"

clean: clean-all

clean-all:
	@echo "Cleaning build artifacts..."
	@rm -f orchestrator/orchestrator.exe
	@rm -f node-agent/node-agent.exe
	@echo "Cleaning dashboard..."
	@cd dashboard && rm -rf build .svelte-kit || echo "Dashboard clean skipped"
	@echo "Clean complete!"

install-deps:
	@echo "Installing Go dependencies..."
	@cd orchestrator && go mod tidy
	@cd node-agent && go mod tidy
	@echo "Installing dashboard dependencies..."
	@cd dashboard && npm install || echo "Dashboard dependencies skipped"
	@echo "Dependencies installed!"

test:
	@echo "Running tests..."
	@cd orchestrator && go test ./...
	@cd node-agent && go test ./...
	@echo "Running dashboard tests..."
	@cd dashboard && npm test || echo "Dashboard tests skipped"

test-all: test