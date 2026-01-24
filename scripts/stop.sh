#!/bin/bash
echo "Stopping servers..."

pkill -f "llm-mux serve" 2>/dev/null && echo "API stopped" || echo "API not running"
pkill -f "vite.*8318" 2>/dev/null && echo "UI stopped" || echo "UI not running"

sleep 1
if lsof -i :8317 &>/dev/null || lsof -i :8318 &>/dev/null; then
  echo "Warning: Some processes still running"
else
  echo "All stopped"
fi
