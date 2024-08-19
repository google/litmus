FROM golang:1.23-alpine AS builder

# Set the working directory in the container
WORKDIR /app

# Explicitly copy the go.mod and go.sum files first
COPY go.mod go.sum ./ 

# Download dependencies 
RUN go mod download

# Copy the rest of your application code
COPY . ./ 

# Build the application
RUN go build -o main .

# Use a smaller image for the final stage
FROM alpine:latest

# Install ca-certificates (required for HTTPS requests)
RUN apk --no-cache add ca-certificates

# Set the working directory in the final image
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Command to run when the container starts
ENTRYPOINT ["./main"]