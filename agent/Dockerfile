FROM golang:1.21 as builder
WORKDIR /app
COPY main.go .
RUN go build -o gpu_agent main.go

FROM debian:bullseye-slim
WORKDIR /app
COPY --from=builder /app/gpu_agent .
COPY config.json .
CMD ["/app/gpu_agent"]
