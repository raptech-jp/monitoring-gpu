FROM nvidia/cuda:12.2.2-runtime-ubuntu22.04 as base
WORKDIR /app
COPY main.go .
RUN apt-get update && apt-get install -y nvidia-utils-525
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o gpu_agent main.go

CMD ["/app/gpu_agent"]
