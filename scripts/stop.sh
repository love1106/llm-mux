#!/bin/bash
echo "Stopping servers..."

pkill -f "llm-mux serve" 2>/dev/null && echo "API stopped" || echo "API not running"
pkill -f "vite.*8318" 2>/dev/null && echo "UI stopped" || echo "UI not running"
pkill -f "scripts/keepalive.sh" 2>/dev/null && echo "Keepalive stopped" || echo "Keepalive not running"

sleep 1

# Force-kill anything still holding the ports
for port in 8317 8318; do
  if fuser "$port/tcp" &>/dev/null 2>&1; then
    echo "Force-killing process on port $port..."
    fuser -k "$port/tcp" 2>/dev/null
  fi
done

if lsof -i :8317 &>/dev/null || lsof -i :8318 &>/dev/null; then
  echo "Warning: Some processes still running"
else
  echo "All stopped"
fi
