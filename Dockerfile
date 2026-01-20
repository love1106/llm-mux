# Stage 1: Build Go server
FROM golang:1.24-alpine AS go-builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w -X 'main.Version=${VERSION}' -X 'main.Commit=${COMMIT}' -X 'main.BuildDate=${BUILD_DATE}'" \
    -o ./llm-mux \
    ./cmd/server/

# Stage 2: Build UI
FROM node:22-alpine AS ui-builder

WORKDIR /build/ui

COPY ui/package.json ui/package-lock.json* ./
RUN npm ci

COPY ui/ ./
RUN npm run build

# Stage 3: Runtime stage
FROM alpine:3.23

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown

LABEL org.opencontainers.image.title="llm-mux" \
      org.opencontainers.image.description="AI Gateway for Subscription-Based LLMs" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.source="https://github.com/love1106/llm-mux"

RUN apk add --no-cache tzdata ca-certificates nginx

RUN addgroup -g 1000 llm-mux && \
    adduser -D -u 1000 -G llm-mux llm-mux && \
    mkdir -p /llm-mux /run/nginx && \
    chown -R llm-mux:llm-mux /llm-mux /run/nginx /var/lib/nginx /var/log/nginx

WORKDIR /llm-mux

COPY --from=go-builder --chown=llm-mux:llm-mux /build/llm-mux ./
COPY --from=ui-builder --chown=llm-mux:llm-mux /build/ui/dist ./ui/dist

COPY <<'EOF' /etc/nginx/http.d/default.conf
server {
    listen 8318;
    root /llm-mux/ui/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /v1/management {
        proxy_pass http://127.0.0.1:8317;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
EOF

COPY <<'EOF' /llm-mux/entrypoint.sh
#!/bin/sh
nginx
exec ./llm-mux serve "$@"
EOF
RUN chmod +x /llm-mux/entrypoint.sh

USER llm-mux
ENV TZ=Asia/Ho_Chi_Minh
EXPOSE 8317
EXPOSE 8318
EXPOSE 54545
ENTRYPOINT ["/llm-mux/entrypoint.sh"]
CMD []
