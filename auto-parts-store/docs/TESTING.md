# Testing Guide

This app has four test layers. Each is independently runnable, and each
tells you something different when it fails.

| Layer | Command | Needs | Where |
|---|---|---|---|
| Unit | `make test-unit` | Nothing | `backend/internal/**/*_test.go`, `web/src/**/*.test.tsx` |
| Integration | `make test-integration` | Docker | `backend/internal/**/*_integration_test.go` |
| API | `make test-api` | Docker | `backend/tests/api/*_test.go` |
| E2E (UI) | `make test-e2e` | Docker | `e2e/*_test.go` |
| Everything | `make test-all` | Docker | all of the above |

Run them in that order when you're not sure where a bug is - unit tests are
fast and isolate business logic; if those pass but integration/API fail,
the bug is almost always in a SQL query, a Pub/Sub wiring issue, or an
HTTP-layer detail (status codes, JSON shape) rather than core logic.

## 1. Unit tests

```sh
make test-unit
```

Runs `go test ./...` in `backend/` and `vitest run` in `web/`. No Docker, no
network - the Go tests run against hand-written in-memory fakes
(`backend/internal/testutil/fakes.go`) implementing the store/pubsub
interfaces, and the frontend tests mock the API client module.

Run just one side:

```sh
cd backend && go test ./...
cd web && npx vitest run
```

Run a single test:

```sh
cd backend && go test ./internal/service/... -run TestCheckoutService_PlaceOrder -v
cd web && npx vitest run src/pages/CheckoutPage.test.tsx
```

## 2. Integration tests

```sh
make test-integration
```

This brings up Postgres + the Pub/Sub emulator via docker-compose, runs
migrations, then runs `go test -tags=integration ./...` in `backend/`.
These tests hit a **real** Postgres (repository CRUD, constraints) and a
**real** Pub/Sub emulator (an end-to-end round trip: publish `OrderPlaced`,
assert the inventory subscriber actually decremented stock in Postgres).

Test files use the `//go:build integration` tag so they're excluded from
plain `go test ./...`. They connect via `testutil.NewTestPostgres(t)` /
`testutil.NewTestPubSubClient(t)`, which each `t.Skip()` if their required
env var (`DATABASE_URL` / `PUBSUB_EMULATOR_HOST`) isn't set - so it's safe
to leave these files in the tree without a Docker daemon running.

Fixture data comes from the seed migration
(`backend/migrations/0002_seed.up.sql`) - tests reference known part
IDs/prices via constants in `testutil.SeedPartOilFilterID` etc. rather than
inserting their own rows, since only `carts`/`orders` (and their line
items) get truncated between test runs
(`Postgres.ResetTransactionalData`) - categories/parts are treated as
static reference data.

## 3. API tests

```sh
make test-api
```

Same infrastructure as integration tests, but black-box: it boots the real
`chi` router wired to the real Postgres + Pub/Sub emulator inside an
`httptest.Server`, then drives it with plain HTTP requests
(`backend/tests/api/helpers_test.go`). This is the layer that catches
router/middleware/status-code/JSON-shape bugs that unit tests can't see
because they never touch HTTP at all.

If you add a new endpoint, add both a unit test for its service-layer logic
*and* an API test asserting the actual HTTP contract (status code + body
shape) - they catch different classes of bugs. A real example from this
codebase: a list endpoint returning an empty Go slice as literal JSON
`null` instead of `[]` was invisible to unit tests (which compare Go
values, not marshaled bytes) but obvious in an API test asserting on the
raw response.

## 4. E2E / UI tests

```sh
make test-e2e
```

This is the heaviest layer: it builds `web` for production, then
`docker compose --profile full up -d --build` brings up **everything**
(Postgres, Pub/Sub emulator, `api`, `worker`, `web` behind nginx), runs
migrations, then runs `go test -tags=e2e ./...` in `e2e/` using
[`playwright-go`](https://github.com/mxschmitt/playwright-go) - same
pattern as this repo's `playwright-example/`. It tears the stack back down
afterward (`docker compose --profile full down -v`), win or lose.

Scenarios covered: browse → view part detail, search filters results,
category filter, add-to-cart → checkout → order confirmation, removing a
cart item updates the total.

Locators use `data-testid` attributes on the React side
(`part-list`, `add-to-cart-{id}`, `cart-total`, `place-order-btn`,
`order-confirmation`, etc. - grep the `web/src` components for the full
list). If you change a component's structure, keep the `data-testid`
stable or update the corresponding e2e test.

### Debugging a failing e2e test

Don't just stare at the Playwright timeout - it tells you *what* didn't
happen, not *why*. Bring the stack up manually and poke at it:

```sh
make build-web
docker compose --profile full up -d --build
make wait-for-postgres
make migrate-up
```

Then either:

- Open http://localhost:5173 in a real browser and click through the flow
  yourself, watching the Network tab.
- Check container logs: `docker compose --profile full logs api` /
  `logs worker` / `logs web`. The `api` logs every request with status code
  and latency - if an expected request never shows up in the log, the bug
  is in the frontend (it never made the call), not the backend.
- Write a throwaway Go script using `playwright-go` that hooks
  `page.On("console", ...)` and `page.On("response", ...)` to print
  everything the browser does - this is how the checkout-redirect race and
  the `null`-vs-`[]` JSON bug in this codebase were actually found. A
  30-second timeout from Playwright only tells you a selector never
  appeared; console/network logging tells you *why*.

Tear down when done:

```sh
docker compose --profile full down -v
```

## Known gotchas

- **First Docker pull is slow.** The Pub/Sub emulator image
  (`gcr.io/google.com/cloudsdktool/cloud-sdk:emulators`) is large. A slow
  first pull is normal; a pull that makes literally zero progress for
  several minutes (check `docker system df` before/after) usually means
  Docker Desktop's registry proxy is stuck - restart Docker Desktop.
- **`make test-e2e` rebuilds images every time.** If you're iterating on a
  frontend fix, `npm run build && docker compose --profile full up -d --build web`
  is faster than the full `make test-e2e` target while you're debugging -
  save the full target for your final check.
- **Stale ports.** If a previous `make test-e2e` run didn't tear down
  cleanly (e.g. you killed the process mid-run), `docker compose --profile
  full down -v` before retrying.
