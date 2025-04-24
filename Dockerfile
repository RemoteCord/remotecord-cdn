# Use the official Golang image as the base image for building
FROM golang:1.22.2 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies (cached if go.mod/go.sum don't change)
RUN go mod download

# Copy the source code into the container
COPY . .

# Ensure the binary is built for Linux (amd64 or arm64 based on your setup)
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go mod tidy && go build -o go-cdn .

# Use a minimal base image for production
FROM alpine:latest

# Install libc6-compat to support Go binaries
RUN apk add --no-cache libc6-compat

# Set the Working Directory inside the container
WORKDIR /root/

# Copy the pre-built binary from the builder stage
COPY --from=builder /app/go-cdn .

# Copy the .env file (if needed)
COPY .env ./

# Ensure the binary has execute permissions
RUN chmod +x go-cdn

# Verify the presence of the binary
RUN ls -la /root

# Expose the port your app runs on
EXPOSE 3002

# Run the application
CMD ["./go-cdn"]