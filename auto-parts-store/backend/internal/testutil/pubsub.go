package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/dl9346/auto-parts-store/backend/internal/pubsub"
)

// NewTestPubSubClient connects to the Pub/Sub emulator pointed to by
// PUBSUB_EMULATOR_HOST (set by `make test-integration` against the
// docker-compose emulator) and ensures the app's topic/subscriptions
// exist. Skips the test if the emulator isn't configured.
func NewTestPubSubClient(t *testing.T) *pubsub.Client {
	t.Helper()

	if os.Getenv("PUBSUB_EMULATOR_HOST") == "" {
		t.Skip("PUBSUB_EMULATOR_HOST not set - run via `make test-integration`")
	}

	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		projectID = "auto-parts-local"
	}

	client, err := pubsub.NewClient(context.Background(), projectID)
	if err != nil {
		t.Fatalf("connect to pubsub emulator: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	if err := client.EnsureTopology(context.Background()); err != nil {
		t.Fatalf("ensure pubsub topology: %v", err)
	}
	return client
}
