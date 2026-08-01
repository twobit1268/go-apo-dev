#!/usr/bin/env bash
# Builds and deploys the three Cloud Run services (api, worker, web) plus
# the Pub/Sub topology, against a real GCP project. Expects a Cloud SQL
# Postgres instance to already exist (Cloud SQL isn't provisioned here -
# see the assumptions in the plan: no Terraform, just these scripts).
#
# Required env vars:
#   GCP_PROJECT_ID        - target GCP project
#   GCP_REGION             - e.g. us-central1
#   CLOUDSQL_CONNECTION_NAME - "project:region:instance"
#   DATABASE_URL           - full Postgres DSN (Cloud SQL Auth Proxy socket)
set -euo pipefail

PROJECT_ID="${GCP_PROJECT_ID:?set GCP_PROJECT_ID}"
REGION="${GCP_REGION:?set GCP_REGION}"
CLOUDSQL_CONNECTION_NAME="${CLOUDSQL_CONNECTION_NAME:?set CLOUDSQL_CONNECTION_NAME}"
DATABASE_URL="${DATABASE_URL:?set DATABASE_URL}"

REPO="auto-parts-store"
AR_HOST="${REGION}-docker.pkg.dev"

gcloud artifacts repositories describe "$REPO" --location="$REGION" --project="$PROJECT_ID" >/dev/null 2>&1 \
  || gcloud artifacts repositories create "$REPO" --repository-format=docker --location="$REGION" --project="$PROJECT_ID"

gcloud auth configure-docker "$AR_HOST" --quiet

API_IMAGE="${AR_HOST}/${PROJECT_ID}/${REPO}/api:$(git rev-parse --short HEAD)"
WORKER_IMAGE="${AR_HOST}/${PROJECT_ID}/${REPO}/worker:$(git rev-parse --short HEAD)"
WEB_IMAGE="${AR_HOST}/${PROJECT_ID}/${REPO}/web:$(git rev-parse --short HEAD)"

docker build -f gcp/cloud-run/api.Dockerfile --build-arg BUILD_TARGET=./cmd/api -t "$API_IMAGE" backend
docker build -f gcp/cloud-run/api.Dockerfile --build-arg BUILD_TARGET=./cmd/worker -t "$WORKER_IMAGE" backend
docker build -f gcp/cloud-run/web.Dockerfile --build-arg VITE_API_BASE_URL="https://api.example.invalid" -t "$WEB_IMAGE" .

docker push "$API_IMAGE"
docker push "$WORKER_IMAGE"
docker push "$WEB_IMAGE"

./gcp/pubsub/create-topics.sh

gcloud run deploy auto-parts-api \
  --project="$PROJECT_ID" --region="$REGION" \
  --image="$API_IMAGE" \
  --add-cloudsql-instances="$CLOUDSQL_CONNECTION_NAME" \
  --set-env-vars="DATABASE_URL=${DATABASE_URL},GCP_PROJECT_ID=${PROJECT_ID}" \
  --allow-unauthenticated

gcloud run deploy auto-parts-worker \
  --project="$PROJECT_ID" --region="$REGION" \
  --image="$WORKER_IMAGE" \
  --add-cloudsql-instances="$CLOUDSQL_CONNECTION_NAME" \
  --set-env-vars="DATABASE_URL=${DATABASE_URL},GCP_PROJECT_ID=${PROJECT_ID}" \
  --no-allow-unauthenticated

gcloud run deploy auto-parts-web \
  --project="$PROJECT_ID" --region="$REGION" \
  --image="$WEB_IMAGE" \
  --allow-unauthenticated

echo "Deployed. Note: rebuild+redeploy auto-parts-web with VITE_API_BASE_URL set to the" \
     "auto-parts-api URL printed above, since it's baked in at build time."
