package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type GPUInfo struct {
	Name           string `json:"name"`
	GPUUtilization string `json:"gpu_utilization"`
	MemoryUsage    string `json:"memory_usage"`
	Temperature    string `json:"temperature"`
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
	var gpus []GPUInfo
	if err := json.NewDecoder(r.Body).Decode(&gpus); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	host := r.RemoteAddr
	s.mu.Lock()
	s.status[host] = gpus
	s.mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	jsonData, err := json.MarshalIndent(s.status, "", "  ")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func (s *Server) displayStatusCLI() {
	for {
		s.mu.Lock()
		fmt.Println("\nGPU Status Report:")
		for host, gpus := range s.status {
			fmt.Printf("Host: %s\n", host)
			for _, gpu := range gpus {
				fmt.Printf(" GPU: %s\n GPU Usage: %s\n Memory Usage: %s\n Temperature: %s\n", gpu.Name, gpu.GPUUtilization, gpu.MemoryUsage, gpu.Temperature)
			}
		}
		s.mu.Unlock()
		fmt.Println("----------------------------------")
		time.Sleep(10 * time.Second)
	}
}

func main() {
	server := NewServer()

	http.HandleFunc("/report", server.reportHandler)
	http.HandleFunc("/status", server.statusHandler)

	fmt.Println("Server started on :8080")
	go server.displayStatusCLI()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
