FROM golang:1.22.5-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o /cert-tracker

# Stage 2: Run
FROM alpine:3.20.2

# Create a group with GID 2000 for fsGroup
RUN addgroup -g 2000 -S fsgroup

# Create a non-root user with specific UID and GID
RUN addgroup -g 3000 -S certgroup && adduser -u 1000 -G certgroup -S certuser

# Set the Current Working Directory inside the container
WORKDIR /home/certuser/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /cert-tracker .

# Change ownership of the working directory to the non-root user
RUN chown -R 1000:2000 /home/certuser/ && chgrp -R 2000 /home/certuser/

# Switch to the non-root user
USER certuser

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./cert-tracker"]