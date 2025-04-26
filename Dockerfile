FROM golang:1.19-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o realtime-quiz-app ./cmd/main.go

# Use a minimal alpine image for the final stage
FROM alpine:latest

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/realtime-quiz-app .
COPY --from=builder /app/migrations ./migrations

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./realtime-quiz-app"]