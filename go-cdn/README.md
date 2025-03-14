# Use the official Golang image as a build stage
FROM golang:1.22.2 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o go-cdn .

# Start a new stage from scratch
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/go-cdn .

# Copy the .env file if needed
COPY .env ./

# Expose port 3002 to the outside world
EXPOSE 3002

# Command to run the executable
CMD ["./go-cdn"]
```

### Explanation:
1. **Base Image**: The Dockerfile starts with the official Golang image to build the application.
2. **Working Directory**: It sets the working directory to `/app`.
3. **Dependency Management**: It copies the `go.mod` and `go.sum` files and runs `go mod download` to install dependencies.
4. **Copy Source Code**: It copies the entire source code into the container.
5. **Build the Application**: It builds the Go application and outputs a binary named `go-cdn`.
6. **Final Stage**: It uses a lightweight Alpine image for the final stage to keep the image size small.
7. **Copy Binary**: It copies the built binary from the builder stage to the final image.
8. **Expose Port**: It exposes port 3002, which is the port your application runs on.
9. **Run Command**: Finally, it specifies the command to run the application.

### Building and Running the Docker Container
To build and run your Docker container, you can use the following commands:

```bash
# Build the Docker image
docker build -t go-cdn .

# Run the Docker container
docker run -p 3002:3002 --env-file .env go-cdn
```

This will build the Docker image and run it, mapping port 3002 of the container to port 3002 on your host machine. Make sure to have your `.env` file in the same directory as your Dockerfile for the environment variables to be loaded correctly.