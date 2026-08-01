// Command api runs the auto parts store's HTTP REST API.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dl9346/auto-parts-store/backend/internal/api"
	"github.com/dl9346/auto-parts-store/backend/internal/config"
	"github.com/dl9346/auto-parts-store/backend/internal/pubsub"
	"github.com/dl9346/auto-parts-store/backend/internal/service"
	"github.com/dl9346/auto-parts-store/backend/internal/store"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pg, err := store.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pg.Close()

	psClient, err := pubsub.NewClient(ctx, cfg.GCPProjectID)
	if err != nil {
		slog.Error("failed to create pubsub client", "error", err)
		os.Exit(1)
	}
	defer psClient.Close()

	if err := psClient.EnsureTopology(ctx); err != nil {
		slog.Error("failed to set up pubsub topology", "error", err)
		os.Exit(1)
	}

	catalogSvc := service.NewCatalogService(pg, pg)
	cartSvc := service.NewCartService(pg, pg)
	checkoutSvc := service.NewCheckoutService(pg, pg, psClient)

	srv := api.NewServer(catalogSvc, cartSvc, checkoutSvc)

	httpServer := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      srv.Router(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("api listening", "port", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}
}
