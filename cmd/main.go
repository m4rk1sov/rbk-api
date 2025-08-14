package main

import (
	"errors"
	"github.com/joho/godotenv"
	"github.com/m4rk1sov/rbk-api/internal/handler"
	"github.com/m4rk1sov/rbk-api/internal/repository"
	"github.com/m4rk1sov/rbk-api/internal/service"
	"github.com/m4rk1sov/rbk-api/pkg/jsonlog"
	"io"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	_ = godotenv.Load()

	addr := getenv("ADDR", ":8080")
	wgerBase := getenv("WGER_BASE_URL", "https://wger.de/api/v2")
	lang := getenvInt("WGER_LANGUAGE", 2)
	ua := getenv("HTTP_USER_AGENT", "rbk-api/1.0 (+https://github.com/m4rk1sov/rbk-api)")
	similarPath := getenv("SIMILAR_MUSCLES_FILE", "./similar_muscles.json")

	logFile, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		log.Printf("failed to open the file for logs: %v", err)
	}
	defer func(logs *os.File) {
		if closeErr := logs.Close(); closeErr != nil {
			log.Printf("failed to close the file for logs: %v", closeErr)
			err = errors.Join(err, closeErr)
		}
	}(logFile)

	logger := jsonlog.New(io.MultiWriter(os.Stdout, logFile), jsonlog.LevelInfo)

	httpClient := &http.Client{Timeout: 10 * time.Second}
	client := repository.NewWgerClient(httpClient, wgerBase, lang, ua)
	svc := service.NewFitnessService(client, logger, similarPath)
	h := handler.New(svc, logger)

	logger.PrintInfo("starting server", map[string]string{"addr": addr})
	if err := http.ListenAndServe(addr, h.Router()); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
