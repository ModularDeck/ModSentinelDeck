# Stage 1: Build the Go binary
FROM golang:latest AS builder

WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application
RUN go build -o sentinel ./main.go

# List files to verify binary creation
RUN ls -la /app

# Stage 2: Create a minimal image to run the compiled binary
FROM alpine:latest

WORKDIR /app

# Copy the compiled Go binary from the builder stage
COPY --from=builder /app/sentinel .

# List files to verify binary is copied
RUN ls -la /app

# Make the binary executable
RUN chmod +x ./sentinel

# Optionally add a version label
LABEL version="1.0.1"

# Expose the port the app runs on
EXPOSE 8080

# Start the application
CMD ["./sentinel"]
