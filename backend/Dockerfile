FROM golang:1.24-alpine3.21

# Set the Current Working Directory inside the container
WORKDIR /src/app

# Install air
RUN go install github.com/air-verse/air@latest

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Download and update dependencies
RUN go mod tidy

