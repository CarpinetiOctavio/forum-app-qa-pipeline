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
