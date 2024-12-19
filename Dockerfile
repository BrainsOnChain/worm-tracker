# Step 1: Build the Go binary
FROM golang:1.22-alpine AS builder

ARG GOARCH=amd64

# Install dependencies
RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code and the .env file
COPY . .

# Ensure the vendor folder is used
RUN go mod vendor

# Build the Go app
RUN CGO_ENABLED=1 GOOS=linux GOARCH=${GOARCH} go build -mod=vendor -o /app/bin/myapp

# Step 2: Create a minimal runtime image
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the Pre-built binary file
COPY --from=builder /app/bin/myapp .

# Command to run the executable
CMD ["./myapp"]
