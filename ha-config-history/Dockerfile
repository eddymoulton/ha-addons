# Stage 1: Build frontend
FROM node:22-alpine AS frontend-builder

WORKDIR /app/frontend

# Copy frontend package files
COPY frontend/package*.json ./

# Install dependencies
RUN npm install

# Copy frontend source
COPY frontend/ ./

# Build frontend (outputs to ../dist)
RUN npm run build

# Stage 2: Build Go application
FROM golang:1.25-alpine AS go-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Build arguments for version information
ARG VERSION=dev
ARG GIT_COMMIT=none
ARG BUILD_TIME=unknown

# Copy Go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Copy built frontend from previous stage
COPY --from=frontend-builder /app/dist ./dist

# Build Go binary with version information
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags "-X main.version=${VERSION} -X main.gitCommit=${GIT_COMMIT}" \
    -o ha-config-history .

# Stage 3: Runtime image
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=go-builder /app/ha-config-history .

# Expose the port
EXPOSE 40613

# Set environment variables
ENV GIN_MODE=release

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:40613/health || exit 1

# Run the application
CMD ["./ha-config-history"]
