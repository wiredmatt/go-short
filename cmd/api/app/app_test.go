package app

import (
	"context"
	"testing"

	"github.com/wiredmatt/go_short/internal/config"
)

func TestNewApp_Success(t *testing.T) {
	ctx := context.Background()
	cfg, err := config.LoadForTest()
	if err != nil {
		panic(err)
	}

	app, err := NewApp(ctx, cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if app.Store == nil {
		t.Fatal("expected non-nil store")
	}
	if app.Server == nil {
		t.Fatal("expected non-nil server")
	}
	if app.Server.Addr != cfg.GetServerAddress() {
		t.Errorf("expected server addr %s, got %s", cfg.GetServerAddress(), app.Server.Addr)
	}
	if app.Server.Handler == nil {
		t.Error("expected non-nil handler")
	}
}

func TestNewApp_InvalidDBType_Error(t *testing.T) {
	ctx := context.Background()
	cfg, err := config.LoadForTest()
	if err != nil {
		panic(err)
	}

	cfg.Database.Type = "somedbtypethatisnotsupported"

	app, err := NewApp(ctx, cfg)
	if err == nil {
		t.Fatal("expected error for invalid db type, got nil")
	}
	if app != nil {
		t.Fatal("expected nil app when error occurs")
	}
}

func TestNewApp_ServerDefaults(t *testing.T) {
	ctx := context.Background()
	cfg, err := config.LoadForTest()
	if err != nil {
		panic(err)
	}

	app, err := NewApp(ctx, cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	server := app.Server
	if server.ReadTimeout != cfg.Server.ReadTimeout {
		t.Errorf("expected ReadTimeout %v, got %v", cfg.Server.ReadTimeout, server.ReadTimeout)
	}
	if server.WriteTimeout != cfg.Server.WriteTimeout {
		t.Errorf("expected WriteTimeout %v, got %v", cfg.Server.WriteTimeout, server.WriteTimeout)
	}
	if server.IdleTimeout != cfg.Server.IdleTimeout {
		t.Errorf("expected IdleTimeout %v, got %v", cfg.Server.IdleTimeout, server.IdleTimeout)
	}
}
