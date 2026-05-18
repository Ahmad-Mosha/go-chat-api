# Build Stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install build dependencies for CGO (SQLite needs it if using mattn/go-sqlite3, though libsql-client-go might not strictly need it, it's safer for SQLite backends)
RUN apk add --no-cache gcc musl-dev

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/chat-api ./cmd/server

# Final Stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/chat-api .
COPY --from=builder /app/.env .env

# Expose port (Render sets this dynamically, but 8080 is our default)
EXPOSE 8080

# Command to run the executable
CMD ["./chat-api"]
