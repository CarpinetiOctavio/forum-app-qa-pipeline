# ADR-003: E2E testing — Cypress

**Date:** 2026-07-28
**Status:** Accepted

## Context

The unit tests inherited from `ci-testing` ([ADR-002](https://github.com/CarpinetiOctavio/forum-app-ci-testing/blob/main/docs/decisions/ADR-002-testing-scope-services-layer.md)) cover the services layer in isolation, against mocked repositories. TP7's scope requires E2E tests that validate complete user flows against the running application — HTTP communication, request/response handling, and UI rendering together, not each in isolation.

## Decision

Use Cypress for E2E testing, with `cy.intercept()` mocks standing in for the backend's HTTP responses rather than a real running backend and database, as the default for every spec **except one**: a small, separate set of non-mocked tests specifically for the `DeleteComment` authorization flow, run against the real backend and a real SQLite database.

`DeleteComment` is a special case because its authorization check lives in the repository's SQL `WHERE ... AND user_id = ?` clause, not in the service layer (see [ADR-005](ADR-005-edit-functionality.md)'s design-constraint section) — the existing unit test for it mocks the repository, so it only proves the service forwards whatever the mock returns; the actual comparison is never exercised by any test in this codebase. A mocked Cypress test for this flow would have the same blind spot for the same reason a mocked unit test does: `cy.intercept()` decides what the "backend" returns, so it can't prove the real SQL clause does the comparison correctly. Every other flow (create, edit, delete-post, error handling) keeps its authorization or validation logic in the service layer, where the existing mocked unit tests already exercise it directly — those don't need a non-mocked Cypress test to close the same gap, because the gap doesn't exist there.

## Rationale

**Cypress over alternatives:**
- Simple, readable syntax with built-in retry logic — reduces the flakiness that comes from manually re-checking async UI state.
- Time-travel debugging and automatic screenshots on failure, which shortens the debugging loop for UI test failures specifically.
- An official GitHub Action, `cypress-io/github-action@v6`, manages the frontend dev server's lifecycle in CI via its own `wait-on` handling, instead of starting the dev server with a backgrounded shell command (`npm start &`) and hand-rolling the health check — the latter is a known source of race conditions (Cypress starts before the dev server is ready) and zombie processes (the backgrounded process outliving the job step), confirmed in `-legacy`'s real implementation (`docs/audits/qa-pipeline-legacy-transferable-knowledge-results.md`).

**Mocks over a real running backend + database, as the default:**
- Tests run in roughly 1–2 minutes instead of 5–7 minutes against a real backend.
- Deterministic results — the test doesn't depend on database state left over from a previous run.
- No race conditions between test runs sharing the same database.

The trade-off is that mocked E2E tests do not validate real backend integration. This is acceptable as the default because: (a) the services layer's logic is already covered by the inherited unit tests, (b) the pipeline separately builds and validates the real backend as its own stage (`backend-build`, already part of the inherited pipeline — see [ADR-004](ADR-004-pipeline-extension.md)), and (c) full integration testing end-to-end is out of this project's scope. It is not an absolute rule: the `DeleteComment` authorization flow is the one case in this codebase where (a) doesn't hold — the unit test covering it doesn't actually exercise the authorization logic (see Decision, above) — so the general trade-off doesn't apply to it, and a real, non-mocked test is used there instead.

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
- **`npm install`, not `npm ci`, in the Cypress CI job.** `npm ci` fails on any drift between `package.json` and `package-lock.json`; `-legacy`'s implementation hit this from adding Cypress as a new dependency and needing the lock file to catch up. Fixed by using `npm install` and committing the resulting lock file, not by hand-editing the lock file (source: `docs/audits/qa-pipeline-legacy-transferable-knowledge-results.md`).
- **`cy.reload()` cannot be used in a way that expects the session to persist.** This app's auth state is transient, held only in React memory (no `localStorage`/`sessionStorage`) — the same model `ci-testing`'s security audit already documented. `cy.reload()` resets the page's JS context, which logs the test user out, the same failure mode that made a test flaky in `-legacy` and had it removed. Specs must account for this (e.g., re-authenticate after a reload, or avoid reloading mid-flow) rather than assume session persistence across a reload.
