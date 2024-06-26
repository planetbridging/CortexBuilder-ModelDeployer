# Start from the latest golang base image
FROM golang:1.22.3

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Expose port 4123 to the outside world
EXPOSE 4125
EXPOSE 12346

# Command to run the Go application
CMD ["go", "run", "."]
