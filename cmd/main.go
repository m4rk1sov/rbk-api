package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/m4rk1sov/rbk-api/internal/handler"
	"log"
	"net/http"
	"os"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("error loading env file: %v", err)
		return
	}
	
	port := os.Getenv("PORT")
	if port == "" {
		log.Println("error loading port from env file, attempting to take a default one...")
		port = "8080"
	}
	
	r := chi.NewRouter()
	r.Get("/fitness/{muscle}", handler.HandleFitness)
	
	log.Println("server running on port:", port)
	log.Fatal(http.ListenAndServe("localhost:"+port, r))
}
