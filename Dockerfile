FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main .

# Production stage
FROM alpine:latest AS production

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/.env .

# Expose the port the app runs on
EXPOSE ${APP_PORT}

# Command to run the application
CMD ["./main"]