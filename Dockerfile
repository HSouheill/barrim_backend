# Dockerfile
FROM golang:1.23

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

RUN go install github.com/air-verse/air@latest

# Copy the source code 
COPY . .

# Build the application
RUN go build -o main .

# Expose the port
EXPOSE 8080

# Run the application
CMD ["air"]