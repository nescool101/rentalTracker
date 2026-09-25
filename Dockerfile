# Multi-stage Dockerfile for React + Go application

# Stage 1: Build React frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend

# Copy package files
COPY frontend/package*.json ./
RUN npm ci

# Copy frontend source and build
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go backend
FROM golang:1.26-alpine AS backend-builder
WORKDIR /app/backend

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source and build
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Stage 3: Final runtime image
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S app && adduser -S -G app app
WORKDIR /app

# Copy the built backend binary
COPY --from=backend-builder /app/backend/main .

# Copy the built frontend static files
COPY --from=frontend-builder /app/frontend/dist ./static

# Writable dirs for runtime-generated signing certs; run as non-root
RUN mkdir -p /app/static /app/certs && chown -R app:app /app
USER app

# Expose port
EXPOSE 8080

# Set environment variables
ENV GIN_MODE=release
ENV SERVER_PORT=8080

# Run the binary
CMD ["./main"] 