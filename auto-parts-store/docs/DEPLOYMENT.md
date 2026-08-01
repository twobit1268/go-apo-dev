# Deploying to GCP

Step-by-step guide to deploy the app to a real GCP project: three Cloud Run
services (`auto-parts-api`, `auto-parts-worker`, `auto-parts-web`) plus a
Pub/Sub topic and two subscriptions, backed by Cloud SQL Postgres.

This app deliberately has no Terraform - it's small enough (two Cloud Run
services doing real work, one static site, one topic) that a couple of
shell scripts are easier to read and modify than a Terraform module. If the
infra grows, that tradeoff should be revisited.

## Prerequisites

- A GCP project with billing enabled
- `gcloud` CLI installed and authenticated (`gcloud auth login`)
- Docker installed locally (the deploy script builds images locally and
  pushes them)
- Owner/Editor-equivalent permissions on the project (Cloud Run, Artifact
  Registry, Cloud SQL, Pub/Sub)

```sh
gcloud config set project YOUR_PROJECT_ID
gcloud services enable run.googleapis.com sqladmin.googleapis.com \
  pubsub.googleapis.com artifactregistry.googleapis.com
```

## 1. Provision Cloud SQL (one-time)

This app doesn't create its own database - point it at an existing Cloud
SQL Postgres instance.

```sh
gcloud sql instances create auto-parts-db \
  --database-version=POSTGRES_16 \
  --tier=db-f1-micro \
  --region=us-central1

gcloud sql databases create autoparts --instance=auto-parts-db

gcloud sql users set-password postgres \
  --instance=auto-parts-db \
  --password=CHOOSE_A_STRONG_PASSWORD
```

Note the instance's connection name (`project:region:instance`):

```sh
gcloud sql instances describe auto-parts-db --format='value(connectionName)'
```

## 2. Run migrations against Cloud SQL

Easiest path: use the [Cloud SQL Auth
Proxy](https://cloud.google.com/sql/docs/postgres/sql-proxy) to reach the
instance from your machine, then run this repo's migration binary against
it.

```sh
cloud-sql-proxy YOUR_PROJECT_ID:us-central1:auto-parts-db &

cd backend
DATABASE_URL="postgres://postgres:CHOOSE_A_STRONG_PASSWORD@localhost:5432/autoparts?sslmode=disable" \
  go run ./cmd/migrate up
```

Re-run this (with `up`) any time you add a new migration file under
`backend/migrations/`.

## 3. Set deploy environment variables

```sh
export GCP_PROJECT_ID=YOUR_PROJECT_ID
export GCP_REGION=us-central1
export CLOUDSQL_CONNECTION_NAME="YOUR_PROJECT_ID:us-central1:auto-parts-db"
export DATABASE_URL="postgres://postgres:CHOOSE_A_STRONG_PASSWORD@localhost/autoparts?host=/cloudsql/${CLOUDSQL_CONNECTION_NAME}&sslmode=disable"
```

The `DATABASE_URL` here uses the Unix-socket form Cloud Run's Cloud SQL
integration expects (`/cloudsql/<connection-name>`), which is different
from the TCP form used against the local proxy in step 2.

## 4. Deploy

```sh
cd auto-parts-store
./gcp/deploy.sh
```

This script (`gcp/deploy.sh`):

1. Creates an Artifact Registry Docker repo if it doesn't exist.
2. Builds and pushes three images (`api`, `worker`, `web`), tagged with the
   current git short SHA.
3. Runs `gcp/pubsub/create-topics.sh` to idempotently create the
   `order-events` topic and its `inventory-sub`/`notifications-sub`
   subscriptions. (The app also creates these itself at startup if they're
   missing - this step just lets a deploy pipeline provision them ahead of
   time without granting the app `pubsub.topics.create`.)
4. Deploys `auto-parts-api` (public) and `auto-parts-worker` (private, no
   public ingress - it only consumes from Pub/Sub) with the Cloud SQL
   connection attached.
5. Deploys `auto-parts-web` (public, static SPA behind nginx).

## 5. Point the web app at the real API

The React app's API base URL is baked in at **build time**
(`VITE_API_BASE_URL`), not read from a runtime env var - the deploy script
builds `web` with a placeholder. After step 4 prints the `auto-parts-api`
URL, rebuild and redeploy just the web image with the real one:

```sh
API_URL=$(gcloud run services describe auto-parts-api \
  --region="$GCP_REGION" --format='value(status.url)')

docker build -f gcp/cloud-run/web.Dockerfile \
  --build-arg VITE_API_BASE_URL="$API_URL" \
  -t "${GCP_REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/auto-parts-store/web:latest" .
docker push "${GCP_REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/auto-parts-store/web:latest"

gcloud run deploy auto-parts-web \
  --project="$GCP_PROJECT_ID" --region="$GCP_REGION" \
  --image="${GCP_REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/auto-parts-store/web:latest" \
  --allow-unauthenticated
```

## 6. Verify

```sh
curl "$API_URL/healthz"        # -> {"status":"ok"}
curl "$API_URL/categories"     # -> seeded categories

WEB_URL=$(gcloud run services describe auto-parts-web \
  --region="$GCP_REGION" --format='value(status.url)')
open "$WEB_URL"                # browse, add to cart, checkout
```

Check the worker actually processed the order:

```sh
gcloud run services logs read auto-parts-worker --region="$GCP_REGION" --limit=20
```

You should see `order confirmation sent` for the order you just placed, and
the part's `stockQty` should have decremented (`curl "$API_URL/parts/<id>"`).

## Redeploying after a code change

Re-run `./gcp/deploy.sh` - it always builds fresh images tagged with the
current commit SHA and redeploys all three services. If you only changed
the frontend, redeploying just `web` (step 5's commands) is faster.

## Rolling back

Cloud Run keeps prior revisions. To roll back a service:

```sh
gcloud run revisions list --service=auto-parts-api --region="$GCP_REGION"
gcloud run services update-traffic auto-parts-api \
  --region="$GCP_REGION" --to-revisions=REVISION_NAME=100
```
