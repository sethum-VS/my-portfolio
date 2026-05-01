# ── Builder Stage ────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

# Install Node.js, npm, and make (often needed for native node modules or esbuild)
RUN apk add --no-cache nodejs npm make git

# Set the working directory
WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

COPY package.json package-lock.json* ./
RUN npm install

# Install templ CLI
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy the rest of the application code
COPY . .

# Generate Templ files
RUN templ generate

# Build Frontend Assets (Tailwind & TypeScript)
# Note: Ensure static/out directory exists or is created by the build scripts
RUN mkdir -p static/out
RUN npm run tailwind:build
RUN npm run build:ts

# Build the Go backend binary
# CGO_ENABLED=0 ensures a statically linked binary, required for scratch/minimal images
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/bin/server ./cmd/server

# ── Runner Stage ─────────────────────────────────────────────────────────────
FROM alpine:latest

# Add ca-certificates for external API calls (e.g., Vertex AI, Firebase)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the compiled Go binary
COPY --from=builder /app/bin/server ./server

# Copy the static assets required by the application
# We copy the entire static directory which now includes static/out (CSS/JS)
COPY --from=builder /app/static ./static

# Expose the application port
EXPOSE 8080

# Run the binary
CMD ["./server"]
