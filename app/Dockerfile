FROM golang:1.23.2-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o /log-generator

# Stage 2: Run
FROM alpine:3.20.3

# Create a group with GID 2000 for fsGroup
RUN addgroup -g 2000 -S fsgroup

# Create a non-root user with specific UID and GID
RUN addgroup -g 3000 -S loggroup && adduser -u 1000 -G loggroup -S loguser

# Set the Current Working Directory inside the container
WORKDIR /home/loguser/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /log-generator .

# Change ownership of the working directory to the non-root user
RUN chown -R 1000:2000 /home/loguser/ && chgrp -R 2000 /home/loguser/

# Switch to the non-root user
USER loguser

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./log-generator"]