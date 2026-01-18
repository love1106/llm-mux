---
title: "Management UI"
description: "React SPA dashboard for llm-mux management"
status: in_progress
priority: P1
effort: 12h
branch: main
tags: [ui, react, dashboard, management]
created: 2026-01-18
---

# llm-mux Management UI Implementation Plan

## Overview
Build a modern React-based management dashboard for llm-mux, embedded in the Go server and served at `/ui/*`. The UI provides comprehensive management capabilities via the existing Management API.

## Architecture
- **Frontend**: React 18 + TypeScript + Vite + Tailwind CSS + shadcn/ui
- **Backend**: Go server with embed.FS serving at port 8317
- **API**: RESTful Management API at `/v1/management/*`
- **Auth**: X-Management-Key header for API authentication

## Core Features
1. **Dashboard**: Overview stats, system health, quick actions
2. **Accounts**: OAuth flow, auth file management, provider status
3. **Usage**: Request/token statistics with time-series charts
4. **Logs**: Real-time log viewer with filtering and search
5. **Settings**: YAML config editor, feature toggles, API keys

## Implementation Phases

### Phase 1: [Project Setup](./phase-01-project-setup.md) (2h) - DONE (2026-01-18)
- React + Vite + TypeScript scaffolding
- Tailwind CSS + shadcn/ui integration
- Development environment configuration

#### Phase 1 Achievements
- React 18 + TypeScript + Vite project initialized
- Tailwind CSS v4 + shadcn/ui configured
- 5 pages: Dashboard, Accounts, Usage, Logs, Settings
- API client with Zustand auth store
- Error boundary and toast notifications
- Build: 116KB gzipped

### Phase 2: [Go Embedding](./phase-02-go-embedding.md) (1h)
- embed.FS integration in Go server
- SPA routing with index.html fallback
- Build pipeline configuration

### Phase 3: [Core Components](./phase-03-core-components.md) (2h)
- Layout structure and navigation
- API client with authentication
- Error handling and loading states

### Phase 4: [Accounts Page](./phase-04-accounts-page.md) (2h)
- Auth file listing and management
- OAuth flow UI (start/status/cancel)
- Provider configuration interface

### Phase 5: [Usage Statistics](./phase-05-usage-stats.md) (2h)
- Time-series charts with Recharts
- Account-level usage breakdown
- Export functionality

### Phase 6: [Logs Viewer](./phase-06-logs-viewer.md) (2h)
- Real-time log streaming with polling
- Advanced filtering and search
- Log level highlighting

### Phase 7: [Settings Page](./phase-07-settings-page.md) (1h)
- YAML config editor with Monaco
- Boolean settings toggles
- API key management

## Success Criteria
- [ ] UI accessible at http://localhost:8317/ui
- [ ] All Management API endpoints integrated
- [ ] Responsive design working on mobile/tablet
- [ ] Real-time updates for logs and metrics
- [ ] Config changes persist correctly
- [ ] OAuth flow completes successfully
- [ ] Build size < 500KB gzipped

## Technical Decisions
- **Routing**: React Router v6 for client-side navigation
- **State**: React Query for server state, Zustand for client state
- **Charts**: Recharts for data visualization
- **Editor**: Monaco Editor for YAML editing
- **Icons**: Lucide React for consistent iconography
- **Forms**: react-hook-form with zod validation

## Research References
- [Vite Embed Research](./research/researcher-01-vite-embed.md)
- [Shadcn Dashboard Research](./research/researcher-02-shadcn-dashboard.md)