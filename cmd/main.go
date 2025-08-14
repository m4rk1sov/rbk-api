package main

import (
	"database/sql"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/m4rk1sov/rbk-api/internal/handler"
	"github.com/m4rk1sov/rbk-api/internal/repository"
	"github.com/m4rk1sov/rbk-api/pkg/jsonlog"
	"io"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
	"os"
	"time"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("failed to open .env file: %v\n", err)
	}

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

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data.db"
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		logger.PrintError("failed to connect to sqlite database", map[string]string{"error": err.Error()})
	}
	defer func(db *sql.DB) {
		if closeErr := db.Close(); closeErr != nil {
			logger.PrintError("failed to close database connection", map[string]string{"error": err.Error()})
			err = errors.Join(err, closeErr)
		}
	}(db)

	repo := repository.NewFitnessRepo(db)
	err = repo.InitTables(db)
	if err != nil {
		logger.PrintError("failed to create tables", map[string]string{"error": err.Error()})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// init router
	r := chi.NewRouter()
	r.Use(repository.RequestLogger(logger))
	r.Use(middleware.Recoverer)

	// routes
	r.Get("/fitness/{muscle}", handler.HandleFitness(repo, logger))
	r.Get("/fitness/", handler.ListAvailable)

	// server
	srv := http.Server{
		Handler:      r,
		Addr:         "localhost:" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err = srv.ListenAndServe(); err != nil {
		logger.PrintFatal("failed to launch the server", map[string]string{"error": err.Error()})
	} else {
		logger.PrintInfo("server running on port: %s", map[string]string{"port": port})
	}
}
