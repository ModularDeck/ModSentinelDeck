# Stage 1: Build the Go binary
FROM golang:latest AS builder

WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application
RUN go build -o sentinel ./cmd/server/main.go
# Use the latest Alpine image for a smaller final image
# Stage 2: Create a minimal image to run the compiled binary
FROM alpine:latest

WORKDIR /app

# Copy the compiled Go binary from the builder stage
COPY --from=builder /app/sentinel .

# Make the binary executable
RUN chmod +x ./sentinel

# Add version label
LABEL version="1.0.0" 