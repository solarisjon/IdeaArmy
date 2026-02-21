# Makefile for AI Agent Team

.PHONY: all build clean test fmt help

# Build all binaries
all: build

# Build all binaries
build:
	@echo "Building all binaries..."
	@mkdir -p bin
	@go build -o bin/cli ./cmd/cli/main.go
	@go build -o bin/cli-v2 ./cmd/cli/main_v2.go
	@go build -o bin/cli-tui ./cmd/cli/main_tui.go
	@go build -o bin/server ./cmd/server/main.go
	@go build -o bin/server-v2 ./cmd/server/main_v2.go
	@echo "✅ Build complete!"
	@ls -lh bin/

# Build individual binaries
cli:
	@go build -o bin/cli ./cmd/cli/main.go

cli-v2:
	@go build -o bin/cli-v2 ./cmd/cli/main_v2.go

cli-tui:
	@go build -o bin/cli-tui ./cmd/cli/main_tui.go

server:
	@go build -o bin/server ./cmd/server/main.go

server-v2:
	@go build -o bin/server-v2 ./cmd/server/main_v2.go

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies installed!"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted!"

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "✅ Vet complete!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f idea_sheet_*.html
	@rm -f example_idea_sheet.html
	@rm -f extended_team_idea_sheet.html
	@rm -f custom_team_idea_sheet.html
	@echo "✅ Clean complete!"

# Run the TUI (requires ANTHROPIC_API_KEY)
run-tui: cli-tui
	@./bin/cli-tui

# Run the v2 CLI (requires ANTHROPIC_API_KEY)
run-cli: cli-v2
	@./bin/cli-v2

# Run the web server
run-server: server-v2
	@./bin/server-v2

# Check if API key is set
check-api-key:
	@if [ -z "$$ANTHROPIC_API_KEY" ] && [ -z "$$ANTHROPIC_KEY" ]; then \
		echo "❌ Error: ANTHROPIC_API_KEY or ANTHROPIC_KEY not set"; \
		echo "Set it with: export ANTHROPIC_API_KEY='your-key'"; \
		exit 1; \
	fi
	@echo "✅ API key is set"

# Full check (fmt + vet + build)
check: fmt vet build
	@echo "✅ All checks passed!"

# Help
help:
	@echo "AI Agent Team - Makefile Commands"
	@echo ""
	@echo "Building:"
	@echo "  make build        - Build all binaries"
	@echo "  make cli          - Build CLI v1"
	@echo "  make cli-v2       - Build CLI v2"
	@echo "  make cli-tui      - Build TUI CLI"
	@echo "  make server       - Build web server v1"
	@echo "  make server-v2    - Build web server v2"
	@echo ""
	@echo "Development:"
	@echo "  make deps         - Install dependencies"
	@echo "  make fmt          - Format code"
	@echo "  make vet          - Run go vet"
	@echo "  make check        - Run fmt + vet + build"
	@echo "  make clean        - Remove build artifacts"
	@echo ""
	@echo "Running:"
	@echo "  make run-tui      - Run TUI (requires API key)"
	@echo "  make run-cli      - Run CLI v2 (requires API key)"
	@echo "  make run-server   - Run web server"
	@echo "  make check-api-key - Check if API key is set"
	@echo ""
	@echo "Other:"
	@echo "  make help         - Show this help message"
