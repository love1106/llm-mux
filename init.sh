#!/bin/bash

mkdir -p auths logs data

if [ ! -f config.yaml ]; then
  cat > config.yaml << 'EOF'
port: 8317
api-keys: []
disable-auth: true
auth-dir: /llm-mux/auth
logging-to-file: true
remote-management:
  allow-remote: false
  secret: ""
usage:
  dsn: "sqlite:///llm-mux/data/usage.db"
  batch-size: 100
  flush-interval: 5s
  retention-days: 30
EOF
fi

chmod 755 auths logs data
chmod 644 config.yaml

echo "Done. Edit config.yaml to configure your settings."
