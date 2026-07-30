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
- **`ci.yml`'s `on:` trigger didn't include direct pushes to `staging` or `main`, only `pull_request`s targeting them and `push`es to `feature/**`.** A merge commit landing on `staging`/`main` — exactly what happens every time a PR merges — triggered nothing: no tests, no coverage gate, no SonarCloud, no Cypress. Confirmed directly, not assumed: `main`'s last real SonarCloud analysis was hours stale, from before this session's Cypress work even started, because nothing had pushed a trigger since. Fixed by adding `staging` and `main` to the `push.branches` list, so a merge to either now runs the full pipeline against what's actually on the branch, not just against the PR's own diff.

  This is the same family of finding as two earlier ones in this repo, worth naming as a pattern rather than treating each occurrence as isolated: (1) `ADR-000` found the inherited `quality-summary` job printed hardcoded pass/fail numbers regardless of what the jobs before it actually did (`if: always()`, no real dependency on their results) — a job whose *output* didn't reflect reality. (2) `docs/rules/testing-and-quality-gates.md`'s branch-protection note: the ruleset only watches `Test Summary` and doesn't auto-track new required jobs, so a job can fail in `ci.yml` without ever blocking a merge — a job that runs and reports honestly, but whose result is never *consulted*. This is a third variant, and a more fundamental one than either: not a job lying about its result, not a result nobody checks, but the pipeline never running at all against the two branches every gate in this repo exists to protect. The common thread across all three: a gate that looks like it's protecting something, verified by nobody, until someone checks the actual mechanism instead of trusting that it must be working.
- **The same `S6505` `allowScripts` fix from the `cypress-e2e` job's `npm install` ([ADR-003](ADR-003-cypress.md)) applies identically to `frontend-tests`' and `frontend-build`'s `npm ci` steps** — same rule, same reason (Node 24's bundled npm doesn't block scripts by default; `allowScripts` needs npm >= 12 to take effect), same fix (`npx --yes npm@12.0.2 ci`). Verified locally before pushing, same as the original case: a clean `npm ci` under npm 12.0.2 blocked the same unrelated packages' scripts and left Cypress's own binary installed and `cypress verify`-clean.
