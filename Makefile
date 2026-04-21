.PHONY: dev install-deps generate tailwind tailwind-build build clean

# ── Development ───────────────────────────────────────────────────────────────
# Starts the Tailwind CSS watcher in the background, then builds and runs the Go server.
dev: generate
	@echo "→ Building initial CSS..."
	@npm run tailwind
	@echo "→ Starting Tailwind CSS watcher (background)..."
	@npm run tailwind:watch &
	@echo "→ Starting JS builder (background)..."
	@npm run watch:ts &
	@echo "→ Building Go server..."
	@mkdir -p ./tmp
	@go build -o ./tmp/main ./cmd/server
	@echo "→ Starting Go server..."
	@./tmp/main

# ── Setup ─────────────────────────────────────────────────────────────────────
# Installs all Go and Node tooling required for development.
install-deps:
	@echo "→ Installing Go CLI tools..."
	@go install github.com/air-verse/air@latest
	@go install github.com/a-h/templ/cmd/templ@latest
	@echo "→ Installing Node packages..."
	@npm install
	@echo ""
	@echo "✓ Done. Ensure $$(go env GOPATH)/bin is in your PATH."

# ── Code Generation ───────────────────────────────────────────────────────────
generate:
	@templ generate

# ── CSS ───────────────────────────────────────────────────────────────────────
tailwind:
	@npm run tailwind

tailwind-build:
	@npm run tailwind:build

ts-build:
	@npm run build:ts

# ── Production Build ──────────────────────────────────────────────────────────
build: generate tailwind-build ts-build
	@mkdir -p ./bin
	@go build -o ./bin/server ./cmd/server
	@echo "✓ Binary at ./bin/server"

# ── Cleanup ───────────────────────────────────────────────────────────────────
clean:
	@rm -rf ./tmp ./bin
	@rm -f ./static/out/output.css
	@echo "✓ Cleaned"
