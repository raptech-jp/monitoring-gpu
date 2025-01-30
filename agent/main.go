package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Config struct {
	ServerAddress string `json:"server_address"`
	ServerPort    string `json:"server_port"`
	FetchInterval int    `json:"fetch_interval"`
}

type GPUInfo struct {
	Name        string `json:"name"`
	MemoryUsage string `json:"memory_usage"`
	MemoryTotal string `json:"memory_total"`
	Temperature string `json:"temperature"`
}

func loadConfig() (*Config, error) {
	file, err := os.ReadFile("config.json")
	if err != nil {
		return nil, err
	}
	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func getGPUStatus() ([]GPUInfo, error) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=name,memory.used,memory.total,temperature.gpu", "--format=csv,noheader,nounits")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	var gpus []GPUInfo
	for _, line := range lines {
		data := strings.Split(line, ", ")
		if len(data) < 4 {
			return nil, fmt.Errorf("unexpected output format")
		}

		gpus = append(gpus, GPUInfo{
			Name:        data[0],
			MemoryUsage: data[1] + "MB",
			MemoryTotal: data[2] + "MB",
			Temperature: data[3] + "°C",
		})
	}
	return gpus, nil
}

func sendToServer(serverAddress, serverPort string, status []GPUInfo) error {
	jsonData, err := json.Marshal(status)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s:%s/report", serverAddress, serverPort)
	fmt.Println("Sending data to:", url)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Failed to send data:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Data sent successfully to", url)
	} else {
		fmt.Println("Failed to send data, server responded with:", resp.Status)
	}

	return nil
}

func main() {
	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	for {
		status, err := getGPUStatus()
		if err != nil {
			fmt.Println("Error fetching GPU status:", err)
		} else {
			sendToServer(config.ServerAddress, config.ServerPort, status)
		}
		time.Sleep(time.Duration(config.FetchInterval) * time.Second) // 設定ファイルの取得間隔を適用
	}
}
