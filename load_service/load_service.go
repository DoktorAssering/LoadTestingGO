package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"
)

type Config struct {
	LoadServicePath string `json:"load_service_path"`
}

var config Config

func LoadConfig() {
	data, err := ioutil.ReadFile("load_service/config.json")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}
}

func StartLoadService() {
	for {
		log.Println("Simulating load...")
		time.Sleep(5 * time.Second)
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	LoadConfig()
}

func main() {
	StartLoadService()
}
