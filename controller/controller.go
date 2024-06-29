package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Config struct {
	LoadServicePath string `json:"load_service_path"`
}

type ReplicaStatus struct {
	ID        string    `json:"id"`
	Host      string    `json:"host"`
	StartTime time.Time `json:"start_time"`
	IsRunning bool      `json:"is_running"`
}

type ScaleRequest struct {
	Action string `json:"action"`
	Count  int    `json:"count"`
}

var (
	mu       sync.Mutex
	replicas = make(map[string]ReplicaStatus)
	hosts    = []string{"localhost:9091", "localhost:9092"}
)

func StartController() {
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/scale", scaleHandler)
	go monitorServices()
	go discoverServices()
	log.Println("Controller started and listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(replicas)
}

func scaleHandler(w http.ResponseWriter, r *http.Request) {
	var req ScaleRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if req.Action == "increase" {
		for i := 0; i < req.Count; i++ {
			go createReplica()
		}
	} else if req.Action == "decrease" {
		for i := 0; i < req.Count; i++ {
			go deleteReplica()
		}
	}
}

func createReplica() {
	for _, host := range hosts {
		resp, err := http.Post("http://"+host+"/agent/create", "application/json", nil)
		if err == nil && resp.StatusCode == http.StatusOK {
			id := fmt.Sprintf("%s-%d", host, time.Now().UnixNano())
			mu.Lock()
			replicas[id] = ReplicaStatus{
				ID:        id,
				Host:      host,
				StartTime: time.Now(),
				IsRunning: true,
			}
			log.Printf("Replica created: %s", id)
			mu.Unlock()
			return
		} else {
			log.Printf("Failed to create replica on host %s: %v", host, err)
		}
	}
}

func deleteReplica() {
	mu.Lock()
	defer mu.Unlock()
	for id, replica := range replicas {
		if replica.IsRunning {
			req, _ := http.NewRequest("DELETE", "http://"+replica.Host+"/agent/delete", nil)
			client := &http.Client{}
			resp, err := client.Do(req)
			if err == nil && resp.StatusCode == http.StatusOK {
				replicas[id] = ReplicaStatus{
					ID:        id,
					Host:      replica.Host,
					StartTime: replica.StartTime,
					IsRunning: false,
				}
				log.Printf("Replica deleted: %s", id)
				return
			} else {
				log.Printf("Failed to delete replica %s on host %s: %v", id, replica.Host, err)
			}
		}
	}
}

func monitorServices() {
	for {
		time.Sleep(10 * time.Second)
		mu.Lock()
		for id, replica := range replicas {
			if replica.IsRunning {
				resp, err := http.Get("http://" + replica.Host + "/agent/status")
				if err != nil || resp.StatusCode != http.StatusOK {
					replicas[id] = ReplicaStatus{
						ID:        id,
						Host:      replica.Host,
						StartTime: replica.StartTime,
						IsRunning: false,
					}
					log.Printf("Replica %s on host %s is not responding", id, replica.Host)
				}
			}
		}
		mu.Unlock()
	}
}

func discoverServices() {
	for {
		time.Sleep(15 * time.Second)
		for _, host := range hosts {
			resp, err := http.Get("http://" + host + "/agent/status")
			if err == nil && resp.StatusCode == http.StatusOK {
				mu.Lock()
				id := fmt.Sprintf("%s-%d", host, time.Now().UnixNano())
				replicas[id] = ReplicaStatus{
					ID:        id,
					Host:      host,
					StartTime: time.Now(),
					IsRunning: true,
				}
				log.Printf("Discovered and added new replica: %s", id)
				mu.Unlock()
			} else {
				log.Printf("Failed to discover service on host %s: %v", host, err)
			}
		}
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	StartController()
}
