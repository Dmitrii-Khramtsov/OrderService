FROM golang:1.24 AS builder
WORKDIR /l0_order/src/orderservice
COPY . .
RUN go mod tidy
RUN go build -o orderservice

FROM debian:bookworm-slim
WORKDIR /l0_order/src/orderservice
COPY --from=builder /l0_order/src/orderservice .
EXPOSE 8080
CMD ["./orderservice"]