# ADR-001: Coverage strategy — 70% threshold, services/components-only scope

**Date:** 2026-07-28
**Status:** Accepted

## Context

TP7's scope adds a coverage gate that blocks the pipeline below a minimum threshold. A coverage percentage is only a meaningful signal if it's measured over code that actually contains business logic — measuring the entire codebase, including HTTP wiring, database setup, and struct definitions, either forces writing low-value tests for trivial code, or produces a precise-looking number that says nothing about whether the business rules are actually verified.

## Decision

Minimum coverage threshold: **70%**, enforced independently in both backend and frontend:
- Backend: a CI step parses `go tool cover` output and fails with a non-zero exit code if the percentage is below 70.
- Frontend: Jest's native `coverageThreshold` in `package.json` (`branches`, `functions`, `lines`, `statements`, all 70).

Coverage is measured only over `backend/internal/services/` and `frontend/src/components/**` + `frontend/src/services/**` — not the rest of the codebase.

## Rationale

**Why 70% and not 100%?**

70% is a pragmatic floor: high enough to catch a real lack of testing discipline, low enough that it doesn't force chasing coverage on defensive branches that can't realistically be triggered or on boilerplate. A 100% requirement tends to produce tests written to satisfy the metric — assertion-free tests that just execute a line — rather than tests that verify behavior.

**Why only `services/` (backend) and `components/` + `services/` (frontend)?**

`backend/internal/services/` is where validations and business rules live — input validation, authorization checks, coordination between repositories. It's the layer with actual branching logic, which is what a coverage gate is meant to protect. The same argument applies to the frontend: components hold the UI logic, and the frontend `services/` layer holds the API-calling logic; everything excluded below is either wiring, generated boilerplate, or type-only code with no behavior to cover.

## Coverage exclusions

**Backend** (`internal/`):

| Package | Reason excluded | How it's tested instead |
|---|---|---|
| `handlers/` | Only maps HTTP request/response to service calls; no business logic of its own | Cypress E2E tests exercise the handlers through real HTTP requests |
| `repository/` | Executes SQL directly against a real connection; not practical to unit-test without a live database | Mocked in every service-layer test, so services are tested against the `Repository` interface, not a concrete implementation |
| `router/` | Route registration only — no conditional logic | Exercised indirectly by every E2E test that hits an endpoint |
| `database/` | Schema definition and connection setup — configuration, not logic | Out of scope for unit tests; would require integration tests against a real SQLite file |
| `models/` | Plain structs with JSON tags — no methods, no logic | N/A — nothing to execute |

Backend `cmd/` (entry point) is validated by the build succeeding, not by coverage.

**Frontend** (`src/`):

| Path | Reason excluded |
|---|---|
| `App.tsx` | Root component — only orchestrates other components and top-level view state; no logic of its own |
| `index.tsx` | React entry point — framework bootstrap |
| `reportWebVitals.ts` | Create React App boilerplate |
| `types/**` | Type/interface definitions only |
| `__mocks__/**` | Test doubles used *by* the tests, not code under test |

## Alternatives considered and rejected

**100% coverage requirement.** Rejected: forces testing of trivial or purely defensive code paths, and incentivizes tests written to satisfy the number rather than to verify behavior.

**Measure the entire codebase, no exclusions.** Rejected: would require integration tests against a real database and HTTP server to meaningfully cover `repository/`, `handlers/`, and `database/` — those layers have no branching logic of their own to protect, so the cost isn't justified by the signal gained.

**A lower threshold (e.g. 50%).** Rejected: too permissive — would pass with large parts of the services layer left untested.

## Consequences

- Backend and frontend coverage scopes are intentionally aligned: `sonar.coverage.exclusions` (see [ADR-002](ADR-002-sonarcloud.md)) and Jest's `collectCoverageFrom` exclude the same categories of non-logic code, so both tools report coverage over comparable ground.
- Because `repository/` is excluded from measurement, any business rule implemented at that layer instead of in `services/` will not be reflected in the coverage number even without a dedicated test for it. This is a known limitation of scoping coverage to `services/` only — the coverage gate alone cannot guarantee every business rule is verified regardless of which layer it happens to live in.
- The Go build (`cmd/`) and the router are validated by the pipeline succeeding, not by a coverage percentage — a compilation failure or a broken route registration surfaces as a failed build/E2E job rather than a coverage regression.
