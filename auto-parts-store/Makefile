.PHONY: dev dev-down migrate-up migrate-down \
        test-unit test-integration test-api test-e2e test-all \
        wait-for-postgres build-web

DATABASE_URL     ?= postgres://postgres:postgres@localhost:5432/autoparts?sslmode=disable
PUBSUB_EMULATOR_HOST ?= localhost:8085
GCP_PROJECT_ID   ?= auto-parts-local

# --- local dev -----------------------------------------------------------

dev:
	docker compose up -d postgres pubsub-emulator
	$(MAKE) wait-for-postgres
	$(MAKE) migrate-up
	@echo ""
	@echo "Postgres + Pub/Sub emulator are up. In separate terminals, run:"
	@echo "  (cd backend && DATABASE_URL='$(DATABASE_URL)' PUBSUB_EMULATOR_HOST=$(PUBSUB_EMULATOR_HOST) GCP_PROJECT_ID=$(GCP_PROJECT_ID) go run ./cmd/api)"
	@echo "  (cd backend && DATABASE_URL='$(DATABASE_URL)' PUBSUB_EMULATOR_HOST=$(PUBSUB_EMULATOR_HOST) GCP_PROJECT_ID=$(GCP_PROJECT_ID) go run ./cmd/worker)"
	@echo "  (cd web && npm run dev)"

dev-down:
	docker compose down -v

wait-for-postgres:
	@until docker compose exec -T postgres pg_isready -U postgres >/dev/null 2>&1; do sleep 1; done

migrate-up:
	cd backend && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate up

migrate-down:
	cd backend && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate down

# --- test pyramid ----------------------------------------------------------

test-unit:
	cd backend && go test ./...
	cd web && npm test -- --run

test-integration:
	docker compose up -d postgres pubsub-emulator
	$(MAKE) wait-for-postgres
	$(MAKE) migrate-up
	cd backend && \
		DATABASE_URL="$(DATABASE_URL)" \
		PUBSUB_EMULATOR_HOST=$(PUBSUB_EMULATOR_HOST) \
		GCP_PROJECT_ID=$(GCP_PROJECT_ID) \
		go test -tags=integration ./...

test-api:
	docker compose up -d postgres pubsub-emulator
	$(MAKE) wait-for-postgres
	$(MAKE) migrate-up
	cd backend && \
		DATABASE_URL="$(DATABASE_URL)" \
		PUBSUB_EMULATOR_HOST=$(PUBSUB_EMULATOR_HOST) \
		GCP_PROJECT_ID=$(GCP_PROJECT_ID) \
		go test -tags=api ./tests/...

test-e2e: build-web
	docker compose --profile full up -d --build
	$(MAKE) wait-for-postgres
	$(MAKE) migrate-up
	cd e2e && WEB_BASE_URL=http://localhost:5173 go test -tags=e2e ./...
	docker compose --profile full down -v

build-web:
	cd web && npm install && npm run build

test-all: test-unit test-integration test-api test-e2e
