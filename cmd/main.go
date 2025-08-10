package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/m4rk1sov/rbk-api/internal/handler"
	"github.com/m4rk1sov/rbk-api/pkg/jsonlog"
	"log"
	"net/http"
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

	logFile := jsonlog.New(f, jsonlog.LevelInfo)
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	err = godotenv.Load(".env")
	if err != nil {
		logFile.PrintError("failed to open env file", map[string]string{"error": err.Error()})
		logger.PrintError("failed to open env file", map[string]string{"error": err.Error()})
	}

	port := os.Getenv("PORT")
	if port == "" {
		logFile.PrintInfo("error loading port from env file, attempting to take a default one...", nil)
		logger.PrintInfo("error loading port from env file, attempting to take a default one...", nil)
		port = "8080"
	}

	r := chi.NewRouter()
	r.Get("/fitness/{muscle}", handler.HandleFitness)

	log.Println("server running on port:", port)
	log.Fatal(http.ListenAndServe("localhost:"+port, r))
}
