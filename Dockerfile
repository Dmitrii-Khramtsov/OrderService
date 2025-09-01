# github.com/Dmitrii-Khramtsov/orderservice/Dockerfile
FROM golang:1.24.1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o orderservice ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/orderservice ./
COPY --from=builder /app/web ./web
COPY --from=builder /app/config.yml ./
EXPOSE 8081
CMD ["./orderservice"]
