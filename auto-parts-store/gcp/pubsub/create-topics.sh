#!/usr/bin/env bash
# Idempotently creates the order-events topic and its two subscriptions
# against a real GCP project. The app also creates these itself at startup
# (internal/pubsub.Client.EnsureTopology) - this script exists for anyone
# who wants topics provisioned ahead of time (e.g. via a deploy pipeline
# that shouldn't grant the app pubsub.topics.create).
set -euo pipefail

PROJECT_ID="${GCP_PROJECT_ID:?set GCP_PROJECT_ID}"
TOPIC="order-events"
SUBSCRIPTIONS=("inventory-sub" "notifications-sub")

if ! gcloud pubsub topics describe "$TOPIC" --project="$PROJECT_ID" >/dev/null 2>&1; then
  echo "creating topic $TOPIC"
  gcloud pubsub topics create "$TOPIC" --project="$PROJECT_ID"
else
  echo "topic $TOPIC already exists"
fi

for sub in "${SUBSCRIPTIONS[@]}"; do
  if ! gcloud pubsub subscriptions describe "$sub" --project="$PROJECT_ID" >/dev/null 2>&1; then
    echo "creating subscription $sub"
    gcloud pubsub subscriptions create "$sub" --topic="$TOPIC" --project="$PROJECT_ID"
  else
    echo "subscription $sub already exists"
  fi
done
