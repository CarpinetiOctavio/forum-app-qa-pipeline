# ADR-004: CI pipeline design — trigger, internal coverage artifact, no coverage gate

**Date:** 2026-07-04
**Status:** Accepted

## Context
`.github/workflows/ci.yml` previously uploaded backend and frontend coverage
output to Codecov, an external SaaS reporting service, via
`codecov/codecov-action`. A decision was needed on whether test automation for
TP6 requires an external coverage-reporting integration, and separately,
whether this pipeline should enforce a minimum coverage percentage as a merge
gate.

## Decision
- **Trigger:** unchanged — runs on `push` and `pull_request` to
  `main`/`master`/`develop`.
- **Jobs:** `backend-tests` and `frontend-tests` run in parallel, each
  building and running its own test suite with coverage instrumentation
  (`go test ./tests/services/... -coverprofile=coverage.out
  -coverpkg=./internal/services/...`; `npm test -- --coverage`).
  `backend-build` and `frontend-build` run after their respective test job
  succeeds; `summary` reports final status.
- **Coverage reporting:** each test job uploads its coverage output
  (`backend/coverage.out`; the `frontend/coverage` directory) as an internal
  GitHub Actions artifact via `actions/upload-artifact`, replacing the
  Codecov integration. No external service is used.
- **No coverage gate:** the pipeline does not fail the build if coverage
  drops below any threshold.

## Rationale
- **A coverage gate is an explicit deliverable of `forum-app-qa-pipeline`
  (TP7), not TP6.** TP6's requirement is that tests run automatically on
  every push and that their results (pass/fail, coverage percentage) are
  visible — it does not require enforcing a minimum. Introducing a gate here
  would silently absorb a requirement that belongs to the next repo in the
  series, undermining the deliberate, incremental scope progression the
  series is built to demonstrate.
- **Removing the external reporting service is justified independently of any
  other repo's configuration.** What TP6's CI needs to prove is that the test
  suite runs automatically and that its coverage output exists and is
  retrievable after the run — both are satisfied by an artifact stored by the
  CI provider itself. An external SaaS integration adds an account/token
  dependency and a third-party surface for exactly zero additional
  capability TP6 requires; it would be justified once coverage needs to be
  tracked and trended over time across many runs, which is a `qa-pipeline`-
  scope concern, not a `ci-testing`-scope one.
- **Parallel test jobs with a build dependency is a correctness ordering, not
  an optimization.** `backend-build`/`frontend-build` depending on their
  test job (`needs: backend-tests` / `needs: frontend-tests`) ensures a
  build is never reported successful on top of a failing test suite —
  the CI's core promise for a testing-automation TP.
- **The backend coverage command must scope to the tested package.** The
  original command, `go test ./... -coverprofile=coverage.out`, has no
  `-coverpkg` flag. Since the test files live in `tests/services` (an
  external test package with no non-test statements of its own), Go
  instruments that package by default and reports `coverage: [no
  statements]` — the artifact this job uploaded was never a real number, in
  any run of this pipeline before this fix. Adding
  `-coverpkg=./internal/services/...` tells Go to attribute coverage to the
  actual package under test (`internal/services`, where `AuthService` and
  `PostService` live), matching the command already documented in
  `docs/COMMANDS.md` for local runs. Verified locally: **54.1%** of
  statements in `internal/services`.

## Alternatives considered
- **Keep Codecov:** rejected per the reasoning above — no TP6 requirement it
  satisfies that an internal artifact does not, at the cost of an external
  dependency.
- **Enforce a minimum coverage percentage now:** rejected — out of TP6's
  scope; see ADR-002 for the same reasoning applied to layer selection.
  Enforcing it here would also require deciding a threshold without the
  broader static-analysis context (`qa-pipeline` pairs its coverage gate with
  SonarCloud), which this repo does not have.
- **Single sequential job instead of parallel backend/frontend jobs:**
  rejected — backend and frontend test suites are fully independent (no
  shared state, no build-order dependency between them), so running them in
  parallel shortens pipeline time with no loss of correctness guarantees.

## Additional corrections
- **Go module rename: `tp06-testing` → `forum-app-ci-testing`.** The former
  name was an identifier assigned by the course for the assignment, not a
  project name. The module in `go.mod`, and every internal import path
  across the backend's `.go` files, was renamed to align with this
  repository's actual name in the portfolio.
- **Go version correction in `ci.yml`.** `go.mod` declared `go 1.24.1` as
  the module's minimum version, but `ci.yml` installed Go 1.21 via
  `actions/setup-go` in both the `backend-tests` and `backend-build` jobs.
  Corrected to `1.24` in both jobs so the CI environment matches what the
  module itself declares as its requirement. Go has handled
  `GOTOOLCHAIN=auto` since 1.21, so the mismatch did not actually break the
  pipeline — but it did contradict what `go.mod` itself states as the
  minimum required version.

## Consequences
- Coverage output is visible per-run as a downloadable artifact but is not
  tracked or trended across runs — that capability, if needed, belongs to
  `qa-pipeline`.
- A pull request can be merged even if coverage drops, as long as tests pass
  — this is a deliberate, scoped decision, not an oversight.
- The backend coverage command bug (uploading a meaningless `[no statements]`
  profile) had gone undetected because nothing in this pipeline previously
  asserted on the *content* of the uploaded artifact — only on whether the
  step succeeded, which it did regardless of what the profile contained.
  There is no coverage gate to catch this class of problem in this repo by
  design (see rationale above); catching a silently-broken measurement
  depends on periodically inspecting what the artifact actually contains,
  which is what surfaced this one.
- **The 54.1% figure above is a point-in-time verification record, not a
  living number.** It documents what was true when this ADR's `-coverpkg`
  fix was verified, the same way `git blame` documents authorship at a
  commit — it is deliberately not updated as the suite changes afterward.
  The actual current coverage differs following ADR-008's security
  hardening (55.2%, one new test added); see the README's Metrics table
  for the current figure, or run the command in `docs/COMMANDS.md`
  directly.

## Branching model correction: from post-hoc verification to preventive gate

### Context
The original trigger (`push` to `main`/`master`/`develop`, plus `pull_request`
to the same branches) let code reach `main` before the pipeline validated it.
Because `push` to `main` was itself a trigger, the pipeline ran *after* the
code was already on the branch it was meant to protect — a failure was
discovered post-hoc, not prevented. A `pull_request` trigger alongside a
`push`-to-`main` trigger is decorative under these conditions: anyone could
bypass the PR entirely by pushing directly to `main`, and the pipeline would
still "pass" retroactively on that same push. This defeats the stated purpose
of a CI gate — validating before merge, not after.

### Decision
Adopt a three-branch model: `feature/* → staging (PR) → main (PR)`.
- `feature/**`: isolated development. Pipeline triggers on `push` — fast
  feedback while the branch is still private to its own line of work.
- `staging`: the first and only path into a shared branch. Pipeline triggers
  on `pull_request` into `staging` and is a required check before merge —
  the first gate.
- `main`: stable, tagged history. Pipeline triggers on `pull_request` into
  `main` and is a required check before merge — the second gate. No trigger
  exists for `push` to `main` at all.

### Rationale
- **Each branch's responsibility is exactly what its trigger enforces, not
  more.** `feature/**` is where iteration happens without any gate on
  merge — a push-triggered run there is diagnostic, not a blocker, because
  nothing merges into a feature branch from outside it. `staging` exists
  specifically to be the first point where a PR — not a push — is what the
  pipeline reacts to; it validates code before it reaches any shared,
  semi-stable branch. `main` repeats that same discipline once more, as the
  last checkpoint before a commit becomes part of the tagged, stable history
  this repo's series treats as a real deliverable.
- **No `develop` branch, deliberately.** A `develop` branch exists in team
  workflows to integrate parallel work from multiple committers before it's
  ready to move forward. This repository has exactly one committer — there
  is no parallel work to integrate, and adding `develop` anyway would have
  meant copying the shape of a team workflow without any of the problem it
  exists to solve. That is worse than not having it: it would look like
  understanding the pattern while actually just imitating its structure.
- **Removing `push` as a trigger for `main` is the actual fix, not a side
  effect.** The single change that converts this pipeline from post-hoc
  verification to a preventive gate is dropping `push: branches: [main]`.
  Everything else in this model (`staging` as an intermediate gate,
  `feature/**` for early feedback) exists to give that removal somewhere to
  land — without an intermediate branch, removing `push` to `main` would
  mean no branch anywhere gets fast, push-triggered feedback during active
  development.

### Alternatives considered
- **`feature/* → develop → staging → main`:** rejected. `develop`'s entire
  reason to exist is reconciling parallel work from more than one committer
  before it's ready to move forward. With a single committer, that
  reconciliation never happens — the branch would carry no real integration
  work, only ceremony. Adding it would have demonstrated the ability to copy
  a team's branching diagram, not the judgment to know which part of it
  applies here.
- **Direct push to `main`, with both `push` and `pull_request` triggers on
  it:** rejected. A `pull_request` trigger is only a gate if the only way to
  reach the protected branch is through a PR. As long as `push` to `main` is
  also a trigger, anyone can skip the PR and push directly — the pipeline
  still runs and still reports green, but after the fact, on code already
  merged. Keeping both triggers on the same branch does not add safety, it
  just adds a second way to reach the same unprotected outcome.

### Consequences
- The commit history on `main` prior to this change is **not rewritten**.
  Every commit before it was made under the old, flawed trigger model — that
  is documented here honestly, as a recognized error corrected going
  forward, not erased or hidden by rewriting history to look as if this
  model was always in place.
- **Commit `d5ff9bc`** (`ci: adopt feature→staging→main branching model`) is
  the exact point where the corrected model takes effect: it changed
  `.github/workflows/ci.yml`'s trigger, updated
  `docs/diagrams/ci-pipeline-flow.svg`, added
  `docs/diagrams/branching-model.svg`, and updated the README's Pipeline &
  Testing section accordingly. Any commit before this one predates the fix
  described in this section.
- **The trigger change alone does not enforce the model.** Nothing in
  `ci.yml` or in Git itself stops a direct `git push origin main` — that
  requires a branch protection rule configured on the GitHub repository
  itself (Settings → Branches: require a pull request before merging,
  require the pipeline as a passing status check), which is **now
  configured on GitHub for `main`** and is out of scope for what a workflow
  file or a local git command can do.
- **PR #1 (`feature/portfolio-setup` → `staging`) did not trigger the
  pipeline as a gate.** GitHub Actions evaluates a workflow using the
  version of it that exists on the PR's base branch — at the time this PR
  was opened, `staging` still carried the old workflow (pre-adoption of this
  model), which had no `pull_request` trigger for `staging` at all. This is
  a known bootstrap problem with any branching-model migration: the first
  PR under a new model is always evaluated against the workflow that
  predates it, not the one it's introducing. Once that PR merged, `staging`
  carried the updated workflow, and the following PR (`staging` → `main`)
  ran correctly as a preventive gate.

## Second correction: the `summary` job never actually checked what it summarized

### Context
While configuring a GitHub ruleset to require status checks on `staging`/
`main` (replacing classic branch protection — see ADR-010 for that
migration itself, which does not change the branching model this ADR
already established, only the GitHub mechanism enforcing it), the obvious
candidate for the required check was `summary` (`name: Test Summary`), the
job that prints the pipeline's final result. Reading the job in full,
not just its name, surfaced a real defect:

```yaml
summary:
  name: Test Summary
  needs: [backend-tests, frontend-tests, backend-build, frontend-build]
  if: always()
  steps:
    - name: Check results
      run: |
        echo "Pipeline Summary:"
        echo "✓ Backend Tests"
        echo "✓ Frontend Tests"
        echo "✓ Backend Build"
        echo "✓ Frontend Build"
```

`if: always()` makes this job run even when a job it depends on failed.
The "Check results" step never inspected the outcome of any of them — four
fixed `echo` lines, no `if` on `needs.*.result`, no conditional `exit`. The
job always succeeded, printing all four checkmarks, regardless of whether
`backend-tests` (or any other dependency) actually failed.

**This is not a regression introduced anywhere in this series.** Verified
with `git log --follow -p -- .github/workflows/ci.yml`: this exact job —
`if: always()` plus the same four unconditional `echo` lines — was
introduced in this repository's very first commit, `7ed7b94` (2025-10-14,
"Proyecto completo: backend + frontend + tests"), as the code the course
handed out for TP6, before this was a portfolio and before the repository
was even named `forum-app-ci-testing`. No commit since — not the Node.js
version bump, not the Actions-version fixes, not the adoption of the
`feature → staging → main` model earlier in this ADR, not the full
portfolio rebuild — ever touched this job. It sat unexamined through every
pass that came after it, until reading it end-to-end for an unrelated
reason (choosing a required check for the ruleset) surfaced it.

**This is the same failure pattern this ADR already documented once, one
layer down, not a coincidence.** The Consequences section above already
describes the backend coverage command silently uploading a meaningless
`[no statements]` artifact for an unknown number of runs, undetected
because nothing checked the *content* of what a successful step produced
— only whether the step itself completed. This is that same audit
question, asked of the layer directly above it: nothing checked the
*result* of the jobs a successful step depended on — only whether the step
itself completed. Both bugs share one root shape: a step reports success
because no real failure case had ever exercised its failure path, not
because the step was actually verified to detect one. The first instance
was found by inspecting what an artifact actually contained instead of
trusting that the step ran. This one was found the same way — by reading
what a job actually did instead of trusting its name and its green
checkmark.

### Decision
Keep `if: always()` — the summary should still run and print its report
even when something upstream failed, for visibility — but make the job's
own outcome depend on its dependencies' real results:

```yaml
summary:
  name: Test Summary
  runs-on: ubuntu-latest
  needs: [backend-tests, frontend-tests, backend-build, frontend-build]
  if: always()

  steps:
    - name: Check results
      run: |
        echo "Pipeline Summary:"
        echo "Backend Tests: ${{ needs.backend-tests.result }}"
        echo "Frontend Tests: ${{ needs.frontend-tests.result }}"
        echo "Backend Build: ${{ needs.backend-build.result }}"
        echo "Frontend Build: ${{ needs.frontend-build.result }}"

    - name: Fail if any dependency failed or was cancelled
      if: contains(needs.*.result, 'failure') || contains(needs.*.result, 'cancelled')
      run: exit 1
```

### Alternatives considered
- **Replace `if: always()` with a job-level condition** (e.g.
  `if: needs.backend-tests.result == 'success' && ...`), letting the job
  itself not run at all when something failed. Rejected: when a job's own
  `if` evaluates to `false`, GitHub Actions marks it **`skipped`**, not
  `failure`. A required status check that reports `skipped` instead of
  `failure` is exactly the kind of ambiguity a ruleset migration
  undertaken for real cost-benefit reasons should not introduce back in at
  the same time — the explicit `exit 1` in the decision above guarantees
  an unambiguous `failure` result regardless of which mechanism
  (classic branch protection or a ruleset) is the one interpreting it.
- **Leave `summary` as an informational-only job and require the four
  individual checks instead** (`Backend Tests (Go)`, `Frontend Tests
  (React)`, `Backend Build`, `Frontend Build`) on the ruleset. This is the
  interim measure already in place while this fix was pending — viable
  long-term too, but it sidesteps the bug rather than fixing it, and a
  single consolidated required check remains simpler to reason about for
  anyone configuring branch protection later. Once this fix lands, whether
  to point the ruleset back at the single `summary` check or keep the four
  individual ones is a separate, later decision.

### Consequences
- If `summary` had been configured as the required check on a branch
  ruleset before this fix, that ruleset would have been decorative — a PR
  with a genuinely broken test suite could still merge, because the check
  being required never reflected the pipeline's real outcome. This was
  caught before any ruleset pointed at `summary` specifically (the interim
  ruleset configuration requires the four individual jobs instead — see
  Alternatives above).
- Verified end to end after this change: backend suite (`go test ./... -v`)
  and frontend suite (`npm test -- --coverage --watchAll=false`) both still
  pass, unaffected by this change (it touches only `.github/workflows/ci.yml`,
  no application code) — see the verification output accompanying this
  commit.
