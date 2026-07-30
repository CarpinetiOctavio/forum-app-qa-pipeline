# ADR-002: Static analysis — SonarCloud

**Date:** 2026-07-28
**Status:** Accepted

## Context

TP7's scope requires a static analysis tool that detects code smells, duplication, and bugs across both Go and TypeScript, integrated with the existing GitHub Actions pipeline, at no cost for a public academic-portfolio repository.

## Decision

Use SonarCloud for static analysis, via `SonarSource/sonarqube-scan-action` as a new CI job.

## Rationale

- Free for public repositories, with no feature restrictions on the quality gate.
- Native GitHub Actions integration — no self-hosted scanner infrastructure needed.
- A single dashboard aggregates both Go and TypeScript results, instead of running two separate linters with two separate reports.
- Coverage reporting integrates directly with the same `coverage.out` / `lcov.info` artifacts already produced by the test jobs (see [ADR-001](ADR-001-coverage-strategy.md)) — no duplicate instrumentation needed.

## Key configuration decisions

**`sonar.language` left unset.** SonarCloud does not support multiple values for this property, and this project is Go + TypeScript. Auto-detection handles both languages correctly without it.

**`sonar.coverage.exclusions` mirrors [ADR-001](ADR-001-coverage-strategy.md)'s exclusion list.** The same layers excluded from `go test`/Jest coverage measurement are excluded here, so SonarCloud's coverage figure and the pipeline's own coverage gate agree on what "coverage" means for this project — a mismatch between the two would be confusing without adding any real signal.

**`sonar.cpd.exclusions` excludes generated coverage artifacts** from the copy-paste detection engine, to avoid false positives on machine-generated report files.

**CI-based analysis, not Automatic Analysis.** SonarCloud's "Automatic Analysis" and CI-based analysis conflict if both are enabled simultaneously. CI-based analysis is the correct mode for a project that already has a pipeline driving every check.

**New Code Definition set to "Number of days: 30", not the "Previous Version" default.** SonarCloud's default Quality Gate includes a "Coverage on New Code" condition, evaluated against whatever the project's New Code Definition considers "new." With the default ("Previous Version"), a commit that only extracts a constant or renames something — with no corresponding new test, because there's no new behavior to test — can register as 0.0% coverage on new code and fail the gate on work that doesn't need a test. This is a known gotcha carried forward from `-legacy`'s real SonarCloud implementation (see `docs/audits/qa-pipeline-legacy-transferable-knowledge-results.md`), not something to rediscover here: the fix is a project-settings change in SonarCloud (New Code Definition → 30 days), not a `sonar-project.properties` change.

## Alternatives considered and rejected

| Alternative | Reason not chosen |
|-------------|------------------|
| CodeClimate | Fewer features available on the free tier |
| Codacy | Free plan is more limited (fewer configurable rules, no full quality gate) |
| ESLint + golangci-lint only | No centralized dashboard, and no single cross-language quality gate — would need to reconcile two separate tools' pass/fail signals manually |

## Consequences

- `SONAR_TOKEN` must be configured as a GitHub Secret in this repository (SonarCloud is a separate service from GitHub Actions, and needs its own credential).
- `sonar-project.properties` must be updated if the repository is ever renamed, since `sonar.projectKey` is tied to the current name.
- The free plan does not allow custom Quality Gates — the SonarCloud job's pass/fail behavior has to work within the default gate's conditions, not a project-specific one.
- **SonarGo doesn't auto-recognize `_test.go` as test code the way SonarJS auto-recognizes `.test.tsx`.** First discovered when `main`'s first real branch analysis (after the `push`-trigger fix) reported `backend/tests/services/post_service_test.go` at 0% new coverage with 143 uncovered lines — the test file itself, not application code. Confirmed against SonarSource's own Go documentation: without `sonar.tests`/`sonar.test.inclusions` set, every line in a `_test.go` file counts as needing coverage and shows 0%, since `coverage.out` never instruments a test file's own lines, only what it exercises. Fixed with `sonar.tests=backend` and `sonar.test.inclusions=backend/**/*_test.go` in `sonar-project.properties` — scoped to `backend/` specifically rather than the docs' generic `sonar.tests=.` example, because this repo mixes `.tsx`/`.test.tsx` in the same frontend directories, which Sonar's own docs flag as exactly the case needing narrower scoping. Frontend test files' `new_lines_to_cover` was 0 before this change (SonarJS's own auto-detection) — checked again against the next real analysis after this config lands, not assumed to still hold just because the change is scoped to `backend/`.
