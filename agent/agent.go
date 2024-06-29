package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

var (
	config   Config
	mu       sync.Mutex
	replicas = make(map[string]ReplicaStatus)
)

func LoadConfig() {
	data, err := ioutil.ReadFile("agent/config.json")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}
}

func StartAgent() {
	http.HandleFunc("/agent/status", statusHandler)
	http.HandleFunc("/agent/create", createHandler)
	http.HandleFunc("/agent/delete", deleteHandler)
	log.Println("Agent started and listening on :9091")
	log.Fatal(http.ListenAndServe(":9091", nil))
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(replicas)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	id := fmt.Sprintf("replica-%d", time.Now().UnixNano())
	replica := ReplicaStatus{
		ID:        id,
		Host:      "localhost:9091",
		StartTime: time.Now(),
		IsRunning: true,
	}
	replicas[id] = replica

	log.Printf("Replica created: %s", id)
	w.WriteHeader(http.StatusOK)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	for id, replica := range replicas {
		if replica.IsRunning {
			replicas[id] = ReplicaStatus{
				ID:        id,
				Host:      replica.Host,
				StartTime: replica.StartTime,
				IsRunning: false,
			}
			log.Printf("Replica deleted: %s", id)
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	LoadConfig()
}

func main() {
	StartAgent()
}
