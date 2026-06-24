# Task Breakdown — src-web-gateway

## Overview
- **Total Phases**: 3
- **Total Tasks**: 8
- **Estimated Total Effort**: M
- **Status**: Phase 1 ✅ | Phase 2 ✅ | Phase 3 ⬜

## S.U.P.E.R Design Constraints

- **S**: Each internal package solves one responsibility (device, server, browser, stream)
- **U**: Data flows: HTTP → Device interface → Browser → CDP; no circular imports
- **P**: Device interface is the contract; server never touches browser directly
- **E**: All config via CLI flags or Config struct; no hardcoded paths
- **R**: ChromeDevice is swappable with future WebRTC/Moonlight backends

## Testing and Governance Constraints

- Tests by default: future feature work must add `*_test.go` files
- Current test exemption: initial scaffolding phase — all verification via `go vet` + `go build`

## Phase 1: Gateway Refactor ✅
**Goal**: Restructure from flat layout to `cmd/` + `internal/` with Device interface + Server layer
**S.U.P.E.R Focus**: P — defining interface contracts before implementing modules

| # | Task | Priority | Effort | Depends | Lane | Status |
|:--|:-----|:---------|:-------|:--------|:-----|:-------|
| G1 | Rewrite `cmd/gateway/main.go` to use `server.New(dev, cfg).Start()` | P0 | S | — | A | ✅ |
| G2 | Write Dockerfile (production mode: Gateway + Chromium) | P1 | S | G1 | A | ✅ |
| G3 | Write docker-compose.yml (debug mode: Chrome + Gateway) | P1 | S | G1 | A | ✅ |
| G4 | Verify `go build`, commit refactor | P0 | S | G2, G3 | A | ✅ |

## Phase 2: Review & Quality Fixes ✅
**Goal**: Fix critical and high-severity issues found by cross-review
**S.U.P.E.R Focus**: E — eliminate platform assumptions; P — enforce interface contracts

| # | Task | Priority | Effort | Depends | Status |
|:--|:-----|:---------|:-------|:--------|:-------|
| G5 | Add `/api/v1/app/stop` route (was missing) | P0 | S | — | ✅ |
| G6 | Fix Swipe duration passthrough (was dead param) | P1 | S | — | ✅ |
| G7 | Fix race conditions (ChromeDevice mutex, Browser atomic.Bool) | P1 | S | — | ✅ |
| G8 | Fix error handling (Tap/Swipe/Key propagation, JSON decode validation) | P1 | S | — | ✅ |
| G9 | Fix HealthStatus Duration → int64 ms, Close idempotency, nil context | P1 | S | — | ✅ |
| G10 | Fix navigate.go context cancellation | P2 | S | — | ✅ |
| G11 | Remove empty dirs, update HANDOFF.md | P2 | S | — | ✅ |

## Phase 3: Production Hardening ⬜
**Goal**: Graceful shutdown, auto-recovery, tests, cookie docs
**S.U.P.E.R Focus**: R — validate replacement test; E — config-driven

| # | Task | Priority | Effort | Depends | Status |
|:--|:-----|:---------|:-------|:--------|:-------|
| G12 | Implement graceful shutdown (SIGINT/SIGTERM handler) | P1 | S | — | ⬜ |
| G13 | Implement Reset() subprocess restart | P1 | M | — | ⬜ |
| G14 | Add Health FSM observer for auto-recovery from DEGRADED | P2 | M | G13 | ⬜ |
| G15 | Add `*_test.go` integration tests | P1 | M | G12 | ⬜ |
| G16 | Document cookie/auth refresh procedure | P2 | S | — | ⬜ |
