package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wiredmatt/go-backend-template/internal/api"
	"github.com/wiredmatt/go-backend-template/internal/config"
	"github.com/wiredmatt/go-backend-template/internal/shortener"
	"github.com/wiredmatt/go-backend-template/internal/storage"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup storage (in-memory for now)
	memStore := storage.NewMemoryStore()

	// Setup service
	shortService := shortener.NewService(memStore, cfg.App.BaseURL, cfg.App.ShortCodeLength)

	// Setup API router (initializes Huma & registers routes)
	router := api.NewRouter(shortService)

	// Create HTTP server with configuration
	server := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting API server on http://%s", cfg.GetServerAddress())
		log.Printf("API docs available at http://%s/docs", cfg.GetServerAddress())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")

}
