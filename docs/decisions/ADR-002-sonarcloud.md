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
- **SonarGo doesn't auto-recognize `_test.go` as test code the way SonarJS auto-recognizes `.test.tsx`.** First discovered when `main`'s first real branch analysis (after the `push`-trigger fix) reported `backend/tests/services/post_service_test.go` at 0% new coverage with 143 uncovered lines — the test file itself, not application code. Confirmed against SonarSource's own Go documentation: without `sonar.tests`/`sonar.test.inclusions` set, every line in a `_test.go` file counts as needing coverage and shows 0%, since `coverage.out` never instruments a test file's own lines, only what it exercises.
  The first fix (`sonar.tests=backend`) was a real scope bug, not an overly-cautious version of the right fix — it broke far more than it repaired. With `sonar.sources` never declared, marking the entire `backend/` tree as the tests root pulled `backend/internal/` and `backend/cmd/` into that root too; neither matched `sonar.test.inclusions`, so both were classified as neither source nor test — invisible to analysis. Confirmed directly, not inferred: for over a day, SonarCloud was analyzing exactly **2 files in the entire project** (`backend/tests/services/auth_service_test.go` and `post_service_test.go`), nothing else in `backend/` and nothing in `frontend/` either, with no project-level `ncloc`/`files`/`coverage` measures at all. Corrected to `sonar.tests=backend/tests` — the actual directory tests and mocks live in — so `backend/internal/` and `backend/cmd/` never enter the tests root at all and fall back to normal source analysis. `sonar.test.inclusions` was widened to `backend/tests/**` (not narrowed further to just `*_test.go`) specifically to avoid orphaning `backend/tests/mocks/*.go` the same way — those files don't match a `_test.go` pattern, so restricting inclusions that tightly would have reproduced the identical bug at smaller scale.
  **The `sonar.tests=backend/tests` correction above was itself incomplete.** After it landed, file count went from 2 to 4 — still only the `backend/tests/services/*_test.go` and `backend/tests/mocks/*.go` files, with `backend/internal/`, `backend/cmd/`, and all of `frontend/` still unanalyzed (`components/search` against SonarCloud's API returned zero results for `backend/internal/services/post_service.go`, a production file that should be ordinary `sonar.sources`). The real CI log's `SonarCloud Scan` step showed the actual cause: `Preprocessing files...` — the file-discovery step, which runs *before* any inclusion/exclusion pattern is applied (`0 files ignored because of inclusion/exclusion patterns`, logged immediately after) — found only 4 files in the entire repository. That rules out `sonar.coverage.exclusions` or `sonar.tests` as the cause of the missing files: discovery itself was already scoped to almost nothing, before any exclusion had a chance to remove anything. `sonar.sources` was never declared in `sonar-project.properties`, on the documented assumption that SonarScanner CLI defaults it to `.` (the whole repo) when unset — that default did not hold in practice here, combined with `sonar.tests` being set, for a reason not clearly documented even in Sonar's own community discussions.
  Fixed by declaring `sonar.sources=backend,frontend` explicitly, removing the ambiguity regardless of the exact mechanism behind the undeclared-default failure. `frontend`, not `frontend/src`, so `frontend/cypress/**` still receives general code-quality analysis — it's excluded from coverage measurement specifically (`sonar.coverage.exclusions`, mirroring ADR-001's Consequences), which is narrower than removing it from analysis entirely.
  To verify once this fix lands: the same real-log check as before (`Preprocessing files...` file count, `Indexing files...` / `N files indexed`), not just the SonarCloud dashboard — the previous fix looked complete from the dashboard's "more than 2 files" framing and wasn't. Pending verification against the next real CI run before this is merged.
