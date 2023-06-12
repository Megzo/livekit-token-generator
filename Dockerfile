# Use the official Go base image
FROM golang:1.20-alpine AS build

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download the Go module dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go application
RUN go build -o server

# Use a minimalistic base image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the binary from the build stage
COPY --from=build /app/server .

# Expose the port that the server listens on
EXPOSE 8080

# Run the binary when the container starts
CMD ["./server"]
