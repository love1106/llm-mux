#!/usr/bin/env bash
set -e

# Install myclaude - Claude Code wrapper for llm-mux

LLM_MUX_URL="${LLM_MUX_URL:-http://localhost:8317}"

# Create Claude Code config to bypass onboarding
mkdir -p ~/.claude

cat > ~/.claude.json << 'EOF'
{
  "hasCompletedOnboarding": true,
  "theme": "dark",
  "numStartups": 1,
  "installMethod": "npm"
}
EOF
echo "Created ~/.claude.json"

# Detect shell config
if [[ -n "$ZSH_VERSION" ]] || [[ "$SHELL" == */zsh ]]; then
  SHELL_RC="$HOME/.zshrc"
else
  SHELL_RC="$HOME/.bashrc"
fi

# Add myclaude function if not exists
if grep -q "^myclaude()" "$SHELL_RC" 2>/dev/null; then
  echo "myclaude already in $SHELL_RC"
else
  cat >> "$SHELL_RC" << EOF

# llm-mux Claude Code wrapper
myclaude() {
  ANTHROPIC_BASE_URL="$LLM_MUX_URL" \\
  ANTHROPIC_API_KEY="sk-ant-api03-mock" \\
  claude "\$@"
}
EOF
  echo "Added myclaude to $SHELL_RC"
fi

echo ""
echo "Run: source $SHELL_RC"
echo "Then: myclaude"
