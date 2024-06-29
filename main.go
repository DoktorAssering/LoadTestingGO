package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

type ReplicaStatus struct {
	ID        string    `json:"id"`
	Host      string    `json:"host"`
	StartTime time.Time `json:"start_time"`
	IsRunning bool      `json:"is_running"`
}

var (
	mu          sync.Mutex
	replicas    = make(map[string]ReplicaStatus)
	hosts       = []string{"localhost:9091", "localhost:9092"}
	replicaBase = 1 // Базовое количество добавляемых реплик
)

func main() {
	logFile, err := os.OpenFile("logs/load_test.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	log.Println("Starting load test...")
	fmt.Println("Starting load test...")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		cmd := exec.Command("go", "run", "controller/controller.go")
		if err := cmd.Start(); err != nil {
			log.Fatalf("Failed to start controller: %v", err)
		}
		log.Printf("controller started with PID %d", cmd.Process.Pid)
		fmt.Printf("controller started with PID %d\n", cmd.Process.Pid)
		cmd.Wait()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		cmd := exec.Command("go", "run", "agent/agent.go")
		if err := cmd.Start(); err != nil {
			log.Fatalf("Failed to start agent1: %v", err)
		}
		log.Printf("agent1 started with PID %d", cmd.Process.Pid)
		fmt.Printf("agent1 started with PID %d\n", cmd.Process.Pid)
		cmd.Wait()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		cmd := exec.Command("go", "run", "agent/agent.go")
		if err := cmd.Start(); err != nil {
			log.Fatalf("Failed to start agent2: %v", err)
		}
		log.Printf("agent2 started with PID %d", cmd.Process.Pid)
		fmt.Printf("agent2 started with PID %d\n", cmd.Process.Pid)
		cmd.Wait()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		cmd := exec.Command("go", "run", "load_service/load_service.go")
		if err := cmd.Start(); err != nil {
			log.Fatalf("Failed to start load service: %v", err)
		}
		log.Printf("load service started with PID %d", cmd.Process.Pid)
		fmt.Printf("load service started with PID %d\n", cmd.Process.Pid)
		cmd.Wait()
	}()

	go func() {
		for i := 1; i <= 20; i++ { // 20 * 30 seconds = 10 minutes
			time.Sleep(30 * time.Second)
			addReplica()
		}
		stopAllReplicas()
	}()

	wg.Wait()
	log.Println("All services have completed.")
	fmt.Println("All services have completed.")
}

func addReplica() {
	mu.Lock()
	defer mu.Unlock()

	replicaCount := len(replicas)
	newReplicaCount := replicaCount + replicaBase
	host := fmt.Sprintf("localhost:%d", 9092+newReplicaCount)
	for i := replicaCount + 1; i <= newReplicaCount; i++ {
		replica := ReplicaStatus{
			ID:        fmt.Sprintf("replica-%d", i),
			Host:      host,
			StartTime: time.Now(),
			IsRunning: true,
		}

		replicas[replica.ID] = replica
		log.Printf("Replica %s started on %s", replica.ID, replica.Host)
		fmt.Printf("Replica %s started on %s\n", replica.ID, replica.Host)
	}
}

func stopAllReplicas() {
	mu.Lock()
	defer mu.Unlock()

	for id, replica := range replicas {
		replica.IsRunning = false
		replicas[id] = replica
		log.Printf("Replica %s stopped", id)
		fmt.Printf("Replica %s stopped\n", id)
	}

	log.Println("Load test completed. Checking results...")
	fmt.Println("Load test completed. Checking results...")

	successfulReplicas := 0
	failedReplicas := 0
	for _, replica := range replicas {
		if replica.IsRunning {
			successfulReplicas++
		} else {
			failedReplicas++
		}
	}

	log.Printf("Load test results: Successful replicas: %d, Failed replicas: %d", successfulReplicas, failedReplicas)
	fmt.Printf("Load test results: Successful replicas: %d, Failed replicas: %d\n", successfulReplicas, failedReplicas)
}
