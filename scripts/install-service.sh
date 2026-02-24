#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
USER_SYSTEMD_DIR="$HOME/.config/systemd/user"

mkdir -p "$USER_SYSTEMD_DIR"

echo "Installing llm-mux user systemd services..."

pkill -f "keepalive.sh" 2>/dev/null || true

cp "$SCRIPT_DIR/llm-mux-api.service" "$USER_SYSTEMD_DIR/"
cp "$SCRIPT_DIR/llm-mux-ui.service" "$USER_SYSTEMD_DIR/"

systemctl --user daemon-reload
systemctl --user enable llm-mux-api llm-mux-ui
systemctl --user start llm-mux-api llm-mux-ui

loginctl enable-linger "$(whoami)" 2>/dev/null || echo "Note: 'loginctl enable-linger' requires sudo for auto-start without login"

echo ""
echo "Done. Services enabled and started."
echo "  API:  systemctl --user status llm-mux-api"
echo "  UI:   systemctl --user status llm-mux-ui"
echo "  Logs: journalctl --user -u llm-mux-api -f"
echo "        journalctl --user -u llm-mux-ui -f"
