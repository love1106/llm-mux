#!/bin/bash
cd "$(dirname "$0")/.."
ROOT="$(pwd)"

if lsof -i :8317 &>/dev/null; then
  echo "API server already running on :8317"
else
  nohup ./llm-mux serve > /tmp/llm-mux-api.log 2>&1 &
  echo "API server started on :8317"
fi

if lsof -i :8318 &>/dev/null; then
  echo "UI server already running on :8318"
else
  (cd ui && nohup npm run dev -- --host 0.0.0.0 > /tmp/llm-mux-ui.log 2>&1 &)
  echo "UI server started on :8318"
fi

pkill -f "keepalive.sh" 2>/dev/null
nohup bash "$ROOT/scripts/keepalive.sh" > /tmp/llm-mux-keepalive.log 2>&1 &
sleep 0.5
if pgrep -f "keepalive.sh" > /dev/null; then
  echo "Keepalive watchdog started"
else
  echo "Warning: Keepalive failed to start"
fi

echo ""
echo "Logs: /tmp/llm-mux-api.log, /tmp/llm-mux-ui.log, /tmp/llm-mux-keepalive.log"
