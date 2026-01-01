# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binaries
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /worker ./cmd/worker
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /scheduler ./cmd/scheduler

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Copy binaries from builder
COPY --from=builder /api /app/api
COPY --from=builder /worker /app/worker
COPY --from=builder /scheduler /app/scheduler

# Copy migrations
COPY --from=builder /app/migrations /app/migrations

# Create logs directory
RUN mkdir -p /app/logs

# Set timezone
ENV TZ=Asia/Ho_Chi_Minh

# Expose port
EXPOSE 8080

# Default command
CMD ["/app/api"]
