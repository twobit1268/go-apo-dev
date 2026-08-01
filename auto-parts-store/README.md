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
UI.

## Docs

- **[docs/GETTING_STARTED.md](./docs/GETTING_STARTED.md)** - clone to
  running app, step by step, plus troubleshooting.
- **[docs/TESTING.md](./docs/TESTING.md)** - how to run and debug each of
  the four test layers.
- **[docs/DEPLOYMENT.md](./docs/DEPLOYMENT.md)** - deploying to a real GCP
  project, step by step.

## Quick reference

```sh
make dev             # start Postgres + Pub/Sub emulator, run migrations
make test-unit        # fast, no Docker
make test-integration # real Postgres + Pub/Sub emulator
make test-api         # real router + Postgres + Pub/Sub emulator, black-box HTTP
make test-e2e          # full docker-compose stack + Playwright-go
make test-all
./gcp/deploy.sh        # deploy to Cloud Run (see docs/DEPLOYMENT.md first)
```
