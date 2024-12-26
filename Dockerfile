# Step 1: Build the Go binary
FROM golang:1.22-alpine AS builder

ARG GOARCH=amd64

# Install dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN CGO_ENABLED=1 GOOS=linux GOARCH=${GOARCH} go build -mod=vendor -o /app/bin/myapp

# Step 2: Create a minimal runtime image
FROM alpine:latest

# Install SQLite runtime
RUN apk add --no-cache sqlite

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the Pre-built binary file
COPY --from=builder /app/bin/myapp .

# Copy the HTML file (or static assets) for runtime
# COPY app/ui.html /root/app/ui.html

# Command to run the executable
CMD ["./myapp"]
