#!/bin/bash
set -e

USER_SYSTEMD_DIR="$HOME/.config/systemd/user"

echo "Removing llm-mux user systemd services..."

systemctl --user stop llm-mux-api llm-mux-ui 2>/dev/null || true
systemctl --user disable llm-mux-api llm-mux-ui 2>/dev/null || true
rm -f "$USER_SYSTEMD_DIR/llm-mux-api.service"
rm -f "$USER_SYSTEMD_DIR/llm-mux-ui.service"
systemctl --user daemon-reload

echo "Done. Services removed."
