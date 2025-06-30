# Build stage
FROM golang:1.23-alpine AS builder

# Set necessary environment variables
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates git

# Create and set the working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -a -installsuffix cgo -ldflags="-w -s" -o bobber ./cmd/main/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user
RUN addgroup -g 1001 -S bobber && \
    adduser -u 1001 -S bobber -G bobber

# Set the working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/bobber .

# Copy configuration files
COPY --from=builder /app/config ./config

# Copy the web directory
COPY --from=builder /app/web ./web

# Change ownership of the app directory
RUN chown -R bobber:bobber /app

# Switch to non-root user
USER bobber

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["./bobber"] 