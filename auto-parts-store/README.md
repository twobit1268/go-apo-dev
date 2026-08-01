# Auto Parts Store

A reference full-stack app demonstrating a Go + GCP + React stack with a
pub/sub-driven REST architecture, plus a complete four-layer test pyramid
(unit, integration, API, UI/e2e).

## Architecture

- **Backend**: Go REST API (`chi` router) backed by Postgres, deployed to
  Cloud Run.
- **Async**: GCP Pub/Sub. Checkout publishes an `OrderPlaced` event on the
  `order-events` topic; two subscribers (run as a separate `worker` process)
  react to it: one decrements inventory, one logs a simulated order
  confirmation.
- **Frontend**: React (Vite) SPA, built as static assets and served
  separately from the API.
- **Local dev/test**: Postgres + the Pub/Sub emulator run in Docker; nothing
  requires a real GCP project until you deploy.

```
auto-parts-store/
  backend/     Go API + worker + migrations
  web/         React SPA
  e2e/         Playwright-go UI tests (separate Go module)
  gcp/         Dockerfiles, deploy script, Pub/Sub setup script
  docker-compose.yml
  Makefile
```

Scope: catalog browsing/search, cart, checkout, order history. No auth (a
bare `customerId` is used), no real payment processor, no admin/inventory
UI - see the plan doc for the full list of assumptions.

## Local development

```sh
make dev        # starts Postgres + Pub/Sub emulator, runs migrations
cd backend && go run ./cmd/api      # in one terminal
cd backend && go run ./cmd/worker   # in another
cd web && npm install && npm run dev
```

The API listens on `:8080`, the web dev server on `:5173`.

## Testing

Four layers, each runnable on its own or all together:

| Layer | Command | What it needs |
|---|---|---|
| Unit | `make test-unit` | Nothing - pure Go/Vitest against in-memory fakes |
| Integration | `make test-integration` | Docker (Postgres + Pub/Sub emulator) |
| API | `make test-api` | Docker (Postgres + Pub/Sub emulator) |
| E2E (UI) | `make test-e2e` | Docker (full stack: Postgres, Pub/Sub emulator, API, worker, web) |
| Everything | `make test-all` | |

- **Unit** (`backend/internal/**/*_test.go`, `web/src/**/*.test.tsx`): table-driven
  tests against hand-written fakes (`backend/internal/testutil`) and mocked
  API clients - no network, no Docker.
- **Integration** (`-tags=integration`): repository tests against a real
  Postgres, plus a Pub/Sub round-trip test (publish `OrderPlaced` on the
  real emulator, assert the inventory subscriber actually decremented
  stock).
- **API** (`-tags=api`, `backend/tests/api`): black-box HTTP tests against
  the real router + real Postgres + real Pub/Sub emulator via
  `httptest.Server`.
- **E2E** (`-tags=e2e`, `e2e/`, using `playwright-go` - same pattern as this
  repo's `playwright-example/`): drives a real Chromium against the full
  docker-compose stack - browse, search/filter, add to cart, checkout,
  confirm the order.

## Deploying to GCP

```sh
export GCP_PROJECT_ID=... GCP_REGION=... CLOUDSQL_CONNECTION_NAME=... DATABASE_URL=...
./gcp/deploy.sh
```

Builds and pushes the `api`, `worker`, and `web` images to Artifact
Registry, provisions the Pub/Sub topic/subscriptions, and deploys three
Cloud Run services. Assumes a Cloud SQL Postgres instance already exists
and migrations have been applied (`make migrate-up` against its
`DATABASE_URL`).
