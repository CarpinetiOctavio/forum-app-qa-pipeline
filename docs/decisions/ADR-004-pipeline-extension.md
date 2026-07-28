# ADR-004: Pipeline extension — adding coverage gating, SonarCloud, and Cypress on top of the inherited pipeline

**Date:** 2026-07-28
**Status:** Accepted

## Context

This repository inherits its CI/CD pipeline from `forum-app-ci-testing@v1.1.0`: `backend-tests`, `frontend-tests`, `backend-build`, `frontend-build`, and a `summary` job that reads the real results of the other four (`needs.*.result`) and fails if any of them failed or were cancelled. The reasoning behind that base structure — why these specific jobs, why parallelized, why a summary job at all — is recorded in `ci-testing`'s own [ADR-004](https://github.com/CarpinetiOctavio/forum-app-ci-testing/blob/main/docs/decisions/ADR-004-ci-pipeline-design.md) and isn't repeated here.

TP7's scope adds three things this base pipeline doesn't have: a coverage threshold gate ([ADR-001](ADR-001-coverage-strategy.md)), static analysis ([ADR-002](ADR-002-sonarcloud.md)), and E2E tests ([ADR-003](ADR-003-cypress.md)). The question this ADR settles is how those get added to the existing pipeline, not whether to add them.

## Decision

Extend the inherited pipeline with two additional jobs, both gated on the existing unit-test jobs:

```
backend-tests ──┐
                ├──> sonarcloud ─────┐
frontend-tests ─┤                    ├──> summary
                └──> cypress-e2e ────┘
backend-build ──────────────────────┘
frontend-build ─────────────────────┘
```

`sonarcloud` and `cypress-e2e` run in parallel with each other, both depending on `backend-tests` and `frontend-tests` passing first (SonarCloud needs the coverage reports those jobs produce; Cypress needs a build that already passed unit tests before spending time on a slower E2E run). The `summary` job's `needs` list is extended to include both new jobs, so a SonarCloud or Cypress failure fails the summary the same way a test or build failure already does.

## Rationale

**Why gate `sonarcloud` and `cypress-e2e` on the test jobs, instead of running everything in parallel from the start?**

Running SonarCloud and Cypress unconditionally in parallel with the test jobs would mean spending CI minutes on static analysis and a 1–2 minute E2E suite even when a unit test is already failing — a signal the pipeline could have surfaced faster by fast-failing on the cheaper jobs first. Gating them on the test jobs passing keeps the fast, cheap signal (unit tests) in front of the slower ones (static analysis, E2E) without giving up parallelization between the two slower stages themselves.

**Coverage gate implementation, concretely:**
- Backend: a step in `backend-tests` parses `go tool cover` output and exits non-zero below 70%.
- Frontend: Jest's native `coverageThreshold` in `package.json` enforces the same threshold natively, without a separate script.

Both live inside the existing `backend-tests`/`frontend-tests` jobs rather than as new standalone jobs — the coverage gate is a property of "did the tests pass with sufficient coverage," not a separate concern from "did the tests pass."

## Alternatives considered and rejected

**Run `sonarcloud` and `cypress-e2e` unconditionally, in parallel with `backend-tests`/`frontend-tests` from the start.** Rejected: burns CI time on static analysis and E2E when a cheaper unit-test failure would have already told the developer the push is broken.

**A dedicated `coverage-gate` job, separate from `backend-tests`/`frontend-tests`.** Rejected: would require re-running the test suite (or passing coverage artifacts between jobs) just to check a threshold that's naturally available at the moment the tests already ran. Keeping the threshold check inline avoids that duplication.

**Make `cypress-e2e` depend on `sonarcloud` (fully sequential) instead of running them in parallel.** Rejected: the two jobs don't depend on each other's output — SonarCloud analyzes source and coverage reports, Cypress drives a running frontend/backend — so serializing them would add wall-clock time for no correctness benefit.

## Consequences

- `summary`'s `needs` list grows from four jobs to six; its "did anything fail" check (`contains(needs.*.result, 'failure')`) doesn't need to change logic, only its dependency list.
- Total pipeline wall-clock time increases (SonarCloud and Cypress both take roughly 1 minute each, run in parallel with each other but after the test jobs) — this is an accepted cost of the added quality gates, not an oversight.
- This ADR only extends the *shape* of the pipeline. The Node-version bump Cypress requires ([ADR-003](ADR-003-cypress.md)) and the `SONAR_TOKEN` secret ([ADR-002](ADR-002-sonarcloud.md)) are prerequisites tracked in their own ADRs, not repeated here.
