FROM golang:1.23.0 AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files, then download dependencies
COPY go.mod go.sum ./

# Copy the rest of the application source code
COPY . .

# Install dependencies and build the application using Makefile
RUN make deps \
    && make build

# Final image
FROM debian:bullseye-slim

# Set environment variables
ENV DATABASE_URL=""

# Set the working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/build/api /usr/local/bin/api

# Expose port 8080
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/api"]

