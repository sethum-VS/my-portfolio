# ── Builder Stage ────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

# Install Node.js, npm, and make (often needed for native node modules or esbuild)
RUN apk add --no-cache nodejs npm make git

# Set the working directory
WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

COPY package.json package-lock.json* ./
RUN npm install

# Install templ CLI at the same version as go.mod (mismatch breaks go build with undefined templ.* symbols)
RUN go install github.com/a-h/templ/cmd/templ@$(go list -m -f '{{.Version}}' github.com/a-h/templ)

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

# ── Runner Stage (Azure Container Apps — port 8080) ─────────────────────────
FROM alpine:latest

# Add ca-certificates for external API calls (e.g., Vertex AI, Firebase)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the compiled Go binary
COPY --from=builder /app/bin/server ./server

# Copy the static assets required by the application
# We copy the entire static directory which now includes static/out (CSS/JS)
COPY --from=builder /app/static ./static

# Port 80 is the target port configured in Azure Container Apps ingress.
# GCP_CREDENTIALS is injected at runtime via Azure Container Apps native secrets
# (stored as secretref:gcp-credentials — never in logs or ARM revision history).
EXPOSE 80

# Run the binary.
# At startup: if GCP_CREDENTIALS is set, write it to a temp file and point
# GOOGLE_APPLICATION_CREDENTIALS at it so the GCP SDK can authenticate.
CMD sh -c 'if [ -n "$GCP_CREDENTIALS" ]; then echo "$GCP_CREDENTIALS" > /tmp/gcp.json; export GOOGLE_APPLICATION_CREDENTIALS=/tmp/gcp.json; fi; exec ./server'
