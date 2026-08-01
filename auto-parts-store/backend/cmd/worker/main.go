// Command worker runs the pub/sub subscribers that react to OrderPlaced
// events: decrementing inventory and sending (simulated) order
// confirmations. Runs as a separate process/Cloud Run service from the API
// so a slow or failing subscriber can't take down the request path.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

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

	inventory := service.NewInventoryHandler(pg)
	notifications := service.NewNotificationHandler()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		slog.Info("inventory subscriber started")
		if err := psClient.Subscribe(ctx, pubsub.InventorySubscription, inventory.Handle); err != nil {
			slog.Error("inventory subscriber stopped", "error", err)
		}
	}()

	go func() {
		defer wg.Done()
		slog.Info("notifications subscriber started")
		if err := psClient.Subscribe(ctx, pubsub.NotificationsSubscription, notifications.Handle); err != nil {
			slog.Error("notifications subscriber stopped", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")
	wg.Wait()
}
