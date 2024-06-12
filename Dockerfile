# Use the official Golang image as a base image
FROM golang:1.22.4 as builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app for linux/arm64
RUN GOOS=linux GOARCH=arm64 go build -o /app/main .

# Final stage
FROM arm64v8/alpine:latest

# Install necessary tools
RUN apk --no-cache add ca-certificates coreutils

# Add a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Switch to the non-root user
USER appuser

# Set the entrypoint
ENTRYPOINT ["./main"]
