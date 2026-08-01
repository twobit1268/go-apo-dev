# Getting Started

Step-by-step instructions to get the auto parts store running on your
machine.

## Prerequisites

- Go 1.26+
- Node.js 22+ and npm
- Docker Desktop (running, not just installed)

Check versions:

```sh
go version
node --version
docker info
```

If `docker info` fails or hangs, start/restart Docker Desktop before
continuing - nothing in `make dev` or the test layers will work without it.

## 1. Clone and orient yourself

```sh
git clone https://github.com/twobit1268/go-apo-dev.git
cd go-apo-dev/auto-parts-store
```

Layout:

```
backend/     Go REST API + worker + migrations
web/         React SPA
e2e/         Playwright-go UI tests (own Go module)
gcp/         Dockerfiles, deploy script, Pub/Sub setup script
docker-compose.yml
Makefile
```

## 2. Start the local infrastructure

```sh
make dev
```

This starts Postgres and the Pub/Sub emulator in Docker and runs the SQL
migrations (schema + seed data: a few sample parts/categories) against
Postgres. It prints the commands you need for step 3 - leave it running.

First run will pull `gcr.io/google.com/cloudsdktool/cloud-sdk:emulators`,
which is a large image. If the pull seems to hang for a long time, that's
usually Docker Desktop's registry proxy being slow, not a broken image -
give it a few minutes before assuming something's wrong.

## 3. Run the API, worker, and web app

In three separate terminals, from `auto-parts-store/`:

```sh
# terminal 1 - REST API on :8080
cd backend && go run ./cmd/api

# terminal 2 - Pub/Sub subscribers (inventory + notifications)
cd backend && go run ./cmd/worker

# terminal 3 - React dev server on :5173
cd web && npm install && npm run dev
```

All three read `DATABASE_URL`, `PUBSUB_EMULATOR_HOST`, and `GCP_PROJECT_ID`
from the environment, defaulting to the docker-compose values from step 2
(see `backend/internal/config/config.go`). You shouldn't need to set
anything by hand for local dev.

## 4. Verify it's working

Open http://localhost:5173. You should see a catalog of seeded auto parts
(brake pads, filters, a battery). Walk through the golden path:

1. Search or filter by category.
2. Click a part, add it to your cart.
3. Go to Cart, then Checkout, then place the order.
4. You should land on an order confirmation page with an order ID and total.

To confirm the async side (Pub/Sub → inventory decrement) actually worked,
check the worker terminal's logs for `order confirmation sent`, and re-check
the part's stock count in the catalog - it should be one lower.

## Troubleshooting

- **`docker info` hangs or `make dev` never finishes pulling images**: Docker
  Desktop isn't running, or its registry proxy is slow/misconfigured. Restart
  Docker Desktop and retry.
- **API can't reach Postgres**: confirm `docker compose ps` shows `postgres`
  as `healthy`, and that nothing else on your machine is bound to `:5432`.
- **Web app loads but API calls fail / CORS errors**: make sure `cmd/api` is
  actually running on `:8080` - the web app has no built-in retry/mock mode.
- **Cart or checkout acts strangely across reloads**: the app has no real
  auth; your "customer" identity and current cart id live in
  `localStorage` (`autoparts.customerId`, `autoparts.cartId`). Clearing
  localStorage resets you to a fresh customer/cart.

## Next steps

- [TESTING.md](./TESTING.md) - running the four test layers
- [DEPLOYMENT.md](./DEPLOYMENT.md) - deploying to GCP
