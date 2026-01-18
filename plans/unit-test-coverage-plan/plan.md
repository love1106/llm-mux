---
title: "Unit Test Implementation Plan for LLM-MUX"
description: "Comprehensive strategy to improve test coverage from 10% to 80% for critical business logic"
status: pending
priority: P1
effort: 12h
branch: main
tags: [testing, quality, go, unit-tests]
created: 2026-01-18
---

# Unit Test Implementation Plan

## Executive Summary
Systematic approach to implement unit tests for llm-mux critical components, targeting 80% coverage
for vital business logic. Focus on provider selection, retry mechanisms, and access control.

## Current State
- **332** source files, **36** test files (~10% coverage)
- All existing tests pass
- Using testify/assert, table-driven patterns

## Implementation Phases

### Phase 1: Provider Core (4h) - **CRITICAL PATH**
**Files:** selector.go, retry.go, quota_manager.go, manager.go
- Round-robin selection with sticky sessions
- Retry/fallback decision logic
- Quota state management
- **Target:** 85% coverage

### Phase 2: Auth & Security (3h)
**Files:** access/manager.go, auth/claude/anthropic_auth.go, oauth/registry.go
- Access control authentication
- OAuth state machine
- PKCE code generation
- **Target:** 80% coverage

### Phase 3: Data & API (3h)
**Files:** usage/sqlite_backend.go, api/handlers/*
- Usage tracking & aggregation
- HTTP handler validation
- **Target:** 70% coverage

### Phase 4: Integration & Benchmarks (2h)
- Race condition testing
- Performance benchmarks
- End-to-end scenarios

## Success Metrics
- [ ] 80% coverage for P0 components
- [ ] All tests pass with `-race`
- [ ] <2s test execution time
- [ ] Zero flaky tests

## Testing Patterns
1. Table-driven tests with descriptive names
2. Interface mocking for external deps
3. Helper functions for setup/teardown
4. Parallel test execution where safe

## Risk Mitigation
- **Risk:** Time constraints → **Mitigation:** Focus on P0 components first
- **Risk:** Complex mocking → **Mitigation:** Use existing test patterns
- **Risk:** Flaky tests → **Mitigation:** Avoid time-dependent tests

## Deliverables
- Test files for each component
- Coverage reports
- CI integration updates
- Documentation updates

## Timeline
- **Day 1:** Phase 1 (Provider Core)
- **Day 2:** Phase 2 (Auth) + Phase 3 (Data)
- **Day 3:** Phase 4 + Documentation