FROM nixos/nix:latest

# Install Go
RUN nix-env -i go

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the application
RUN go build -o main .

# Run the application
CMD ["./main"]
