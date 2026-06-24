# Risk Assessment — src-web-gateway

## S.U.P.E.R Architecture Health Summary

| Principle | Status | Key Findings | Priority |
|:----------|:-------|:-------------|:---------|
| **S** Single Purpose | 🟡 | `server.go` mixes HTTP routing and capture loop; `browser` mixes CDP, nav, and cookies | Low |
| **U** Unidirectional Flow | 🟢 | Clear directional flow: SRC → HTTP → Device interface → Browser → CDP → Chrome | — |
| **P** Ports over Implementation | 🟡 | Device interface well-defined; compat dismiss violates interface via type-assertion | Medium |
| **E** Environment-Agnostic | 🟡 | Mobile UA hardcoded; Chrome flags platform-specific; game URL hardcoded | Low |
| **R** Replaceable Parts | 🟢 | Device interface enables clean backend swapping; browser layer replaceable | — |

**Overall Health**: _3/5 principles healthy_ — Healthy with minor technical debt

### S.U.P.E.R Violation Hotspots
1. `internal/server/server.go:266` — Compat dismiss type-asserts to `*device.ChromeDevice`, bypassing Device interface
2. `internal/browser/browser.go` — Hardcoded mobile UA and stealth JS strings
3. `internal/browser/navigate.go:15` — Hardcoded game URL (`sr.mihoyo.com/cloud/`)

## Risk Matrix

| Risk | Impact | Likelihood | Severity | Mitigation |
|:-----|:-------|:-----------|:---------|:-----------|
| Chrome process crash with no auto-recovery | High | Medium | High | Implement Reset() subprocess restart (deferred) |
| Cookie expiration breaks login | High | Low | Medium | Document cookie refresh procedure |
| Cloud game DOM changes break navigation | High | Medium | Medium | Navigation uses CSS selectors; monitor upstream |
| No automated tests | Medium | High | Medium | Add integration tests with Docker Compose |
| Race conditions on device state | Low | Low | Low | Fixed — sync.Mutex on ChromeDevice, atomic.Bool on Browser.alive |

## Technical Debt
- `Reset()` is a stub — closes browser but doesn't restart
- No graceful shutdown (SIGINT/SIGTERM handler)
- Health FSM not closed-loop (no observer to auto-recover from DEGRADED)
- Compat API aliases are unused by current SRC adapter (dead code)
- `InspectWhite()` never called (dead code in navigate.go)
- Screenshot ignores Format field (always JPEG, never PNG)

## Testing Risks
- Zero `*_test.go` files — no unit or integration tests
- No CI pipeline
- Manual verification only: `go build` + Docker Compose smoke test
- End-to-end validation requires working Chrome container + cloud game auth

## Project Governance Risks
- No AGENTS.md or CLAUDE.md — future contributors lack project-specific instructions
- HANDOFF.md is session-specific; not intended as permanent docs
- No documented cookie/auth refresh procedure
