# Forum App — QA Pipeline

**Author:** Octavio Carpineti
**Course:** Software Engineering III — Universidad Católica de Córdoba (UCC)
**Year:** 2025

This is the second repository in a three-part series, each one adding exactly one
layer of complexity on top of a foundation that is already fundamented and
closed: [forum-app-ci-testing](https://github.com/CarpinetiOctavio/forum-app-ci-testing) → **forum-app-qa-pipeline** (this repo) → [forum-app-cloud-deploy](https://github.com/CarpinetiOctavio/forum-app-cloud-deploy).
For the reasoning behind why this series is split into three repos instead of
one, see ci-testing's [Why This Repository Exists](https://github.com/CarpinetiOctavio/forum-app-ci-testing#why-this-repository-exists) — it isn't repeated here.

This repository is built as a copy of `forum-app-ci-testing@v1.1.0`, not a
continuation of `forum-app-qa-pipeline-legacy` (the original TP7 submission,
now archived). See [`ADR-000`](docs/decisions/ADR-000-starting-from-ci-testing-v1.1.0.md)
for why.

---

## Scope

Building on the unit-testing foundation `forum-app-ci-testing` already
established, this repository's declared scope (TP7) is:

- Code coverage measurement with a 70% gate
- Static analysis via SonarCloud
- End-to-end testing via Cypress
- Quality gates that block the pipeline on failure

This is a statement of intent, not a status report. What's actually implemented
as of this writing:

| Piece | Status |
|---|---|
| Backend + frontend unit tests (inherited from `ci-testing`) | ✅ In place |
| CI pipeline (unit tests, build) | ✅ In place |
| Coverage gate (70%) | ⏳ Not yet implemented |
| SonarCloud static analysis | ⏳ Not yet implemented |
| Cypress E2E tests | ⏳ Not yet implemented |
| Quality gate blocking the pipeline | ⏳ Not yet implemented |

Each row moves to "in place" in the same commit or PR that implements it — see
`docs/rules/documentation.md`.

---

## Tech Stack

| Layer | Technology |
|-------|------------|
| Backend | Go 1.24 + SQLite |
| Frontend | React 18 + TypeScript |
| Unit testing | Go testing + Jest |
| E2E testing (planned) | Cypress |
| Static analysis (planned) | SonarCloud |
| CI/CD | GitHub Actions, Node 24 |

Node is on 24, not the 18 inherited from `ci-testing` — Cypress 13+ requires
20+, and 24 is the current active LTS rather than a version already past
end-of-life. `ci-testing` stays on 18 for its own scope, which has no Cypress
dependency; see the relevant ADR for this repo's own reasoning.

---

## Getting Started

See [`docs/SETUP.md`](docs/SETUP.md) for infrastructure setup and
[`docs/COMMANDS.md`](docs/COMMANDS.md) for the commands that work against this
repo's current state.

---

## Documentation

- `docs/decisions/` — ADRs for decisions made in this repo. Decisions this repo
  builds on from `ci-testing` are linked, not duplicated (see
  `docs/rules/documentation.md`).
- `docs/rules/` — operating rules for anyone (human or AI) working in this repo.
- `docs/audits/` — audit reports that informed decisions recorded in this
  repo's ADRs, kept as evidence, not summarized away.

---

## AI Usage Disclosure

Claude acted as a conceptual auditor and writing assistant during this
repository's setup — never as decision-maker for scope, testing strategy, or
what to carry forward from `forum-app-qa-pipeline-legacy`. All design decisions
were made by Octavio Carpineti; Claude's role was surfacing inconsistencies,
verifying claims against the actual repository and its predecessors, and
drafting documentation for review. See `CLAUDE.md` for the operating rules that
govern this.