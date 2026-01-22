.PHONY: build test clean release help docs dev

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X 'main.Version=$(VERSION)' -X 'main.Commit=$(COMMIT)' -X 'main.BuildDate=$(DATE)'

build:
	@CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o llm-mux ./cmd/server/

dev: build
	@./llm-mux serve

test:
	@go test ./...

clean:
	@rm -rf llm-mux dist/

release:
	@./scripts/release.sh $(filter-out $@,$(MAKECMDGOALS)) || true

%:
	@:

help:
	@echo "make build            - Build binary"
	@echo "make dev              - Build and start server"
	@echo "make test             - Run tests"
	@echo "make clean            - Remove artifacts"
	@echo "make release [cmd]    - Run release script"
	@echo "make docs             - Deploy docs to gh-pages"

docs:
	@echo "📚 Deploying docs to gh-pages..."
	@git push origin $$(git subtree split --prefix docs):gh-pages --force
	@echo "✅ Docs deployed to https://nghyane.github.io/llm-mux/"
