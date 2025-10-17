# Build stage
FROM golang:1.25-alpine AS builder

# Set necessary environment variables
ENV CGO_ENABLED=0 \
  GOOS=linux

# Install git for go mod download
RUN apk add --no-cache git

# Set the working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -a -installsuffix cgo -o main ./cmd/immichcjobs

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Set the working directory
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Create a directory for the job state file
RUN mkdir -p /data

# Set the working directory to /data for runtime
WORKDIR /data

# Copy the binary to /usr/local/bin for easier access
COPY --from=builder /app/main /usr/local/bin/immichcjobs

# Set environment variables with default values (non-sensitive only)
ENV IMMICH_API_URL="localhost:2283" \
  CRON_EXPRESSION="0 * * * *" \
  LAST_CREATED_DIR="/data"

# Run the application
CMD ["/usr/local/bin/immichcjobs"]
