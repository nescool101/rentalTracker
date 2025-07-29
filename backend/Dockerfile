# Build Stage
ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm AS builder

# Ensure CA certificates are available in the build stage
RUN apt update && apt install -y ca-certificates

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /run-app .

# Final Stage
FROM debian:bookworm

# Install CA certificates and curl for healthchecks
RUN apt update && apt install -y ca-certificates curl && update-ca-certificates

COPY --from=builder /run-app /usr/local/bin/
CMD ["run-app"]