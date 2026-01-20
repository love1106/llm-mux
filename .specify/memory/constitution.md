<!--
Sync Impact Report:
- Version: 1.0.0 (initial)
- Added: All sections (initial constitution)
- Ratification: 2026-01-20
- Templates requiring updates: N/A (initial)
-->

# Project Constitution

**Project**: llm-mux
**Version**: 1.0.0
**Ratification Date**: 2026-01-20
**Last Amended**: 2026-01-20

## Purpose

This constitution defines the non-negotiable principles, development workflows, and governance rules for the llm-mux project.

## Principles

### Principle 1: Development Server Commands

**Rule**: When the user asks to "start server", "start servers", or "run the server":
- MUST start the Go API backend directly (e.g., `go run . serve` or `./llm-mux serve`)
- MUST start the Admin UI dev server (e.g., `cd ui && npm run dev`)
- MUST NOT use Docker Compose or Docker containers unless explicitly requested

**Rationale**: Local development benefits from hot-reload and faster iteration. Docker is for deployment, not daily development.

### Principle 2: Code Quality

**Rule**: All code changes MUST:
- Pass existing tests
- Follow existing code style and conventions
- Not introduce new linting errors

**Rationale**: Maintain codebase consistency and prevent regressions.

### Principle 3: Configuration Priority

**Rule**: Configuration sources are loaded with this priority (highest to lowest):
1. Environment variables
2. Command-line flags
3. Configuration file
4. Default values

**Rationale**: Standard 12-factor app pattern for flexible deployment.

## Governance

### Amendment Procedure

1. Propose changes with rationale
2. Update version according to semver:
   - MAJOR: Breaking governance changes
   - MINOR: New principles or sections
   - PATCH: Clarifications and fixes
3. Update `Last Amended` date
4. Propagate changes to dependent templates

### Compliance Review

Constitution principles SHOULD be reviewed when:
- Adding new features that affect development workflow
- Changing project structure
- Onboarding new contributors
