#!/bin/bash
ROOT="$(dirname "$0")/.."
cd "$ROOT"

echo "Watchdog started. Monitoring ports 8317 (API) and 8318 (UI)..."

while true; do
  # Check and restart API server if needed
  if ! lsof -i :8317 &>/dev/null; then
    echo "[$(date)] API server down, restarting..."
    nohup ./llm-mux serve > /tmp/llm-mux-api.log 2>&1 &
  fi

  # Check and restart UI server if needed
  if ! lsof -i :8318 &>/dev/null; then
    echo "[$(date)] UI server down, restarting..."
    (cd ui && nohup npm run dev -- --host 0.0.0.0 > /tmp/llm-mux-ui.log 2>&1 &)
  fi

  sleep 5
done
