package main

import (
	"github.com/joho/godotenv"
	"github.com/m4rk1sov/rbk-api/pkg/jsonlog"
	"log"
	"os"
)

func main() {
	f, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		log.Printf("failed to open the file for logs: %v", err)
	}
	defer func(logs *os.File) {
		if err = logs.Close(); err != nil {
			log.Printf("failed to close the file for logs: %v", err)
		}
	}(f)
	
	logger := jsonlog.New(f, jsonlog.LevelInfo)
	
	err = godotenv.Load(".env")
	if err != nil {
		logger.PrintError("failed to open env file", map[string]string{"error": err.Error()})
	}
	key := os.Getenv("API_KEY")
	//https://wger.de/api/v2/
}
