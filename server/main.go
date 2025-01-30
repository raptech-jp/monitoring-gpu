package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type GPUInfo struct {
	Name        string `json:"name"`
	MemoryUsage string `json:"memory_usage"`
	MemoryTotal string `json:"memory_total"`
	Temperature string `json:"temperature"`
}

type Server struct {
	mu     sync.Mutex
	status map[string][]GPUInfo
}

func NewServer() *Server {
	return &Server{
		status: make(map[string][]GPUInfo),
	}
}

func (s *Server) reportHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request from:", r.RemoteAddr)

	var gpus []GPUInfo
	if err := json.NewDecoder(r.Body).Decode(&gpus); err != nil {
		fmt.Println("Failed to decode JSON:", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	fmt.Println("Received GPU data:", gpus)

	host := r.RemoteAddr
	s.mu.Lock()
	s.status[host] = gpus
	s.mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	processedStatus := make(map[string][]map[string]string)
	for host, gpus := range s.status {
		var gpuDetails []map[string]string
		for _, gpu := range gpus {
			used, err1 := strconv.Atoi(gpu.MemoryUsage[:len(gpu.MemoryUsage)-2])
			total, err2 := strconv.Atoi(gpu.MemoryTotal[:len(gpu.MemoryTotal)-2])
			usagePct := "N/A"
			if err1 == nil && err2 == nil && total > 0 {
				usagePct = fmt.Sprintf("%.2f%%", (float64(used)/float64(total))*100)
			}

			gpuDetails = append(gpuDetails, map[string]string{
				"name":          gpu.Name,
				"memory_usage":  gpu.MemoryUsage,
				"memory_total":  gpu.MemoryTotal,
				"memory_usage_pct": usagePct,
				"temperature":   gpu.Temperature,
			})
		}
		processedStatus[host] = gpuDetails
	}

	jsonData, err := json.MarshalIndent(processedStatus, "", "  ")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func main() {
	server := NewServer()

	http.HandleFunc("/report", server.reportHandler)
	http.HandleFunc("/status", server.statusHandler)

	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
