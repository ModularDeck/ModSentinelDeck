# Stage 1: Build the Go binary
FROM golang:1.23.0 AS builder

WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sentinel ./cmd/server/main.go

# Stage 2: Create a minimal image to run the compiled binary
FROM alpine:latest

WORKDIR /app

# Install CA certificates (optional but often needed for HTTPS)
RUN apk --no-cache add ca-certificates

# Copy the statically linked Go binary
COPY --from=builder /app/sentinel .

# Make sure it's executable
RUN chmod +x ./sentinel

# Optionally add a version label
LABEL version="1.0.1"

# Expose the port the app runs on
EXPOSE 8080

# Run the binary
ENTRYPOINT ["./sentinel"]