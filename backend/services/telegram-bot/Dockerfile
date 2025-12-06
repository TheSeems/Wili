# Build stage
FROM golang:1.24.5-alpine AS builder

WORKDIR /app

# Copy workspace files
COPY go.work go.work.sum ./

# Copy all workspace modules (required by go.work)
COPY devutil/ ./devutil/
COPY services/user/ ./services/user/
COPY services/wishlist/ ./services/wishlist/
COPY services/telegram-bot/ ./services/telegram-bot/

# Set working directory to telegram bot
WORKDIR /app/services/telegram-bot

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/services/telegram-bot/main .

# Expose port
EXPOSE 8080

# Command to run
CMD ["./main"]

