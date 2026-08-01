package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/dl9346/auto-parts-store/backend/internal/store"
)

// NewTestPostgres connects to DATABASE_URL (set by `make test-integration` /
// `make test-api` against the docker-compose Postgres) and truncates every
// table so each test starts from a clean slate. It skips the test (rather
// than failing) if DATABASE_URL isn't set, so `go test ./...` without the
// integration/api build tags never needs a database.
func NewTestPostgres(t *testing.T) *store.Postgres {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set - run via `make test-integration` or `make test-api`")
	}

	pg, err := store.NewPostgres(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect to postgres: %v", err)
	}
	t.Cleanup(pg.Close)

	// Only transactional data is reset - categories/parts come from the
	// migration seed data (0002_seed.up.sql) and tests reference those
	// fixture IDs/prices directly rather than inserting their own.
	if err := pg.ResetTransactionalData(context.Background()); err != nil {
		t.Fatalf("reset transactional data: %v", err)
	}
	return pg
}

// Known fixture IDs/prices from backend/migrations/0002_seed.up.sql, for
// integration and API tests to reference without re-deriving them.
const (
	SeedPartBrakePadID     = "aaaaaaaa-0001-0001-0001-000000000001" // price 4899
	SeedPartOilFilterID    = "aaaaaaaa-0002-0002-0002-000000000001" // price 899
	SeedPartBrakePadPrice  = 4899
	SeedPartOilFilterPrice = 899
)
