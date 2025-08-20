package main

import (
	"context"
	"net/http"

	"github.com/wiredmatt/go-backend-template/internal/api"
	"github.com/wiredmatt/go-backend-template/internal/config"
	"github.com/wiredmatt/go-backend-template/internal/shortener"
	"github.com/wiredmatt/go-backend-template/internal/storage"
)

type App struct {
	Cfg    *config.Config
	Store  storage.Store
	Server *http.Server
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	store, err := storage.NewStore(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}

	shortService := shortener.NewService(store, cfg.App.BaseURL, cfg.App.ShortCodeLength)
	router := api.NewRouter(shortService)

	server := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &App{
		Cfg:    cfg,
		Store:  store,
		Server: server,
	}, nil
}
