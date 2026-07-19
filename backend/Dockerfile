# Multi-stage build for Go backend
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o inkless-api ./cmd/server

# Final stage - minimal runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates wget tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/inkless-api .

# Expose port
EXPOSE 8088

# Run the server
CMD ["./inkless-api"]
