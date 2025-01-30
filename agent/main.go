package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

type GPUStatus struct {
	GPUUtilization string `json:"gpu_utilization"`
	MemoryUsage    string `json:"memory_usage"`
	Temperature    string `json:"temperature"`
}

func getGPUStatus() (*GPUStatus, error) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=utilization.gpu,memory.used,temperature.gpu", "--format=csv,noheader,nounits")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	data := strings.Split(strings.TrimSpace(out.String()), ", ")
	if len(data) < 3 {
		return nil, fmt.Errorf("unexpected output format")
	}

	return &GPUStatus{
		GPUUtilization: data[0] + "%",
		MemoryUsage:    data[1] + "MB",
		Temperature:    data[2] + "°C",
	}, nil
}

func sendToServer(status *GPUStatus) error {
	jsonData, err := json.Marshal(status)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://your-server-ip:8080/report", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func main() {
	for {
		status, err := getGPUStatus()
		if err != nil {
			fmt.Println("Error fetching GPU status:", err)
		} else {
			sendToServer(status)
		}
		time.Sleep(10 * time.Second) // 10秒ごとに取得
	}
}
