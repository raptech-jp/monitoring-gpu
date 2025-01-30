FROM golang:1.21 as builder
WORKDIR /app
COPY main.go .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o gpu_server main.go  # 静的ビルド

FROM debian:bullseye-slim
WORKDIR /app
COPY --from=builder /app/gpu_server .
CMD ["/app/gpu_server"]
