package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/wiredmatt/go-backend-template/internal/api"
	"github.com/wiredmatt/go-backend-template/internal/shortener"
	"github.com/wiredmatt/go-backend-template/internal/storage"
)

func main() {
	// Setup storage (in-memory for now)
	memStore := storage.NewMemoryStore()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		log.Fatal("BASE_URL environment variable is not set")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Setup service
	shortService := shortener.NewService(memStore, baseURL)

	// Setup API router
	router := api.NewRouter(shortService)

	log.Println("Starting API on :" + port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}
