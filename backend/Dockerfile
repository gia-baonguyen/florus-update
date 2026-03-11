# Build stage
FROM golang:1.24rc1-alpine AS builder

ENV GOTOOLCHAIN=auto

WORKDIR /app

# Install dependencies
RUN apk add --no-cache gcc musl-dev git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application (GOTOOLCHAIN already set at top)
RUN GOTOOLCHAIN=auto CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/.env.example .env

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
