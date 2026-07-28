# ADR-003: E2E testing — Cypress

**Date:** 2026-07-28
**Status:** Accepted

## Context

The unit tests inherited from `ci-testing` ([ADR-002](https://github.com/CarpinetiOctavio/forum-app-ci-testing/blob/main/docs/decisions/ADR-002-testing-scope-services-layer.md)) cover the services layer in isolation, against mocked repositories. TP7's scope requires E2E tests that validate complete user flows against the running application — HTTP communication, request/response handling, and UI rendering together, not each in isolation.

## Decision

Use Cypress for E2E testing, with `cy.intercept()` mocks standing in for the backend's HTTP responses rather than a real running backend and database.

## Rationale

**Cypress over alternatives:**
- Simple, readable syntax with built-in retry logic — reduces the flakiness that comes from manually re-checking async UI state.
- Time-travel debugging and automatic screenshots on failure, which shortens the debugging loop for UI test failures specifically.
- An official GitHub Action (`cypress-io/github-action`) manages the frontend dev server's lifecycle in CI, instead of hand-rolling process start/stop/health-check logic in a shell step.

**Mocks over a real running backend + database:**
- Tests run in roughly 1–2 minutes instead of 5–7 minutes against a real backend.
- Deterministic results — the test doesn't depend on database state left over from a previous run.
- No race conditions between test runs sharing the same database.

The trade-off is that mocked E2E tests do not validate real backend integration. This is acceptable here because: (a) the services layer's logic is already covered by the inherited unit tests, (b) the pipeline separately builds and validates the real backend as its own stage (`backend-build`, already part of the inherited pipeline — see [ADR-004](ADR-004-pipeline-extension.md)), and (c) full integration testing end-to-end is out of this project's scope.

## Alternatives considered and rejected

| Alternative | Reason not chosen |
|-------------|------------------|
| Playwright | More complex configuration surface; multi-browser support isn't needed for this project's scope |
| Selenium | Heavier setup; well-documented history of flaky tests in comparable projects |
| Puppeteer | Chrome-only; a lower-level API that requires more boilerplate for the same assertions Cypress provides natively |

## Consequences

- **Node version dependency:** Cypress 13+ requires Node 20 or newer. This repo's current CI configuration (`.github/workflows/ci.yml`, inherited from `ci-testing@v1.1.0`) runs the backend/frontend test and build jobs on Node 18. Adopting Cypress means bumping the Node version used by those existing jobs as a prerequisite of this change, not an independent, separately-scheduled upgrade.
- `cypress.config` must define `baseUrl` pointing at the frontend dev server.
- E2E tests depend on the frontend being available at `localhost:3000` when Cypress runs; the CI job needs `wait-on` (or equivalent) to avoid racing the dev server's startup.
