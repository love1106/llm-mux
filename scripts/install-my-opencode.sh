#!/usr/bin/env bash
set -e

LLM_MUX_URL="${LLM_MUX_URL:-http://10.154.149.74:8317}"
[[ "${LLM_MUX_URL}" != */ ]] && LLM_MUX_URL="${LLM_MUX_URL}/"

mkdir -p ~/.config/opencode
rm -f ~/.config/opencode/opencode.json

cat > ~/.config/opencode/opencode.json << 'EOF'
{
  "$schema": "https://opencode.ai/config.json",
  "plugin": ["oh-my-opencode@latest"],
  "provider": {
    "anthropic": {
        "options": {
          "baseURL": "{env:ANTHROPIC_BASE_URL}/v1",
          "apiKey": "{env:ANTHROPIC_API_KEY}"
      }
    }
  },
  "permission": "allow",
  "autoupdate": true
}
EOF
echo "Created ~/.config/opencode/opencode.json"

if [[ -n "$ZSH_VERSION" ]] || [[ "$SHELL" == */zsh ]]; then
  SHELL_RC="$HOME/.zshrc"
else
  SHELL_RC="$HOME/.bashrc"
fi

sed -i '/^myopencode()/,/^}/d' "$SHELL_RC" 2>/dev/null || true

cat >> "$SHELL_RC" << EOF

myopencode() {
  ANTHROPIC_BASE_URL="$LLM_MUX_URL" \\
  ANTHROPIC_API_KEY="sk-ant-api03-mock" \\
  opencode "\$@"
}
EOF
echo "Added myopencode to $SHELL_RC"

echo ""
echo "Run: source $SHELL_RC"
echo "Then: myopencode"
echo ""
echo "Debug: OPENCODE_DEV_DEBUG=true myopencode"
