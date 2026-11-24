# ADR-006: SQLite as the database engine — retroactive record, and a known defect: foreign keys declared but not enforced

**Date:** 2026-11-23
**Status:** Accepted

## Context

The application has used SQLite since the original TP6 submission, and this
repository inherited it with the rest of the code ([ADR-000](ADR-000-starting-from-ci-testing-v1.1.0.md)).
No decision record in this series states that choice or its reasons.

**When the reasons were written.** The earliest versioned record is
`backend/internal/database/desc.md` in the original TP6, added in
[`forum-app-ci-testing@938577a`](https://github.com/CarpinetiOctavio/forum-app-ci-testing/blob/938577a/backend/internal/database/desc.md)
(2025-10-16). The original TP7 repeated them in section 1.3 of
[`decisiones.md`](https://github.com/CarpinetiOctavio/forum-app-qa-pipeline-legacy/blob/296fad6/decisiones.md)
(`forum-app-qa-pipeline-legacy`, as restored to its October 2025 state).

**How they were lost.** Two separate steps dropped them:

- The 2026 rebuild of `forum-app-ci-testing`
  ([`a7dd683`](https://github.com/CarpinetiOctavio/forum-app-ci-testing/commit/a7dd683))
  removed every `desc.md` and `decisiones.md`. Its stack decision, `ci-testing`'s
  [ADR-001](https://github.com/CarpinetiOctavio/forum-app-ci-testing/blob/main/docs/decisions/ADR-001-stack-choice.md),
  covers Go and React and says nothing about the database.
- This repository's audit of `-legacy`
  ([`qa-pipeline-legacy-audit-results.md`](../audits/qa-pipeline-legacy-audit-results.md), section 4)
  discarded the "why SQLite" part of `database/desc.md` because it overlapped
  with `-legacy`'s ADR-001
  ([`889101c`](https://github.com/CarpinetiOctavio/forum-app-qa-pipeline-legacy/blob/889101c/docs/decisions/ADR-001-tech-stack.md)),
  and `-legacy`'s ADRs were not carried over ([ADR-000](ADR-000-starting-from-ci-testing-v1.1.0.md)).
  The same audit described the `ON DELETE CASCADE` as "confirmed in code": it
  confirmed the declaration, not that SQLite applies it.

**The defect.** The schema in `backend/internal/database/database.go`
declares three foreign keys with `ON DELETE CASCADE`: `posts.user_id` and
`comments.user_id` referencing `users`, and `comments.post_id` referencing
`posts`. SQLite enforces foreign keys only on connections that enable them;
they are off by default
([SQLite, "Enabling Foreign Key Support"](https://www.sqlite.org/foreignkeys.html#fk_enable)).
`InitDB("./database.db")` does not enable them: the connection string has no
`_foreign_keys` or `_fk` parameter, and no `PRAGMA foreign_keys = ON` is
issued. The driver, `github.com/mattn/go-sqlite3` v1.14.32, sets that pragma
only when the connection string asks for it. As a result, no foreign key is
checked:

- `DeletePost` runs `DELETE FROM posts WHERE id = ?`. The post's comments stay
  in `comments`, pointing to a post that no longer exists. They are not shown,
  because comments are only read through their post (`WHERE c.post_id = ?`).
- The database accepts a row that references a missing one.

The defect was verified by reading `database.go`, `cmd/api/main.go` and the
driver's source, and by running the same kind of schema in the `sqlite3`
command-line shell: `PRAGMA foreign_keys` returns `0`, deleting a post leaves
its comment, and with `PRAGMA foreign_keys = ON` the comment is deleted. It
was not reproduced by running this application, because the local toolchain
used for the check could not link the cgo driver. No test covers it: the
service-layer tests mock the repositories. `database.go` and the `InitDB`
call are identical in `forum-app-ci-testing`, where the defect originates.

## Decision

1. **SQLite is recorded as the database engine**, with the reasons given when
   the repository was created. This is a retroactive record of a decision made
   in TP6 and inherited here, not a new choice.
2. **The foreign-key defect is recorded and is not fixed in this repository.**
   The documentation states what the code does: the cascade is declared, not
   enforced.

## Reasons recorded at the time

From the records cited above:

- No separate server process: the database is a single file, created
  automatically on first run.
- Simple local development and CI setup, without Docker or additional
  configuration.
- Sufficient for the academic scope of the project.

`decisiones.md` also cites a clean database on every test run. That reason
does not apply to this repository: the unit tests never open the database.

## Other engines: rejected at the time, and considered today

**PostgreSQL or MySQL.** Rejected because they need a running server, make
local development and the CI configuration more complex, and were judged
excessive for the scope of the assignment; the TP6 record also notes them as
better suited to production. PostgreSQL enforces foreign keys and their
cascades on every connection, without configuration, so it would have removed
the possibility of the current defect. Requiring a separate server to run it
would still have oversized a course project.

**DuckDB, considered on 2026-11-23 (not recorded at the time).** It would not
be used today either, because it does not fit this use case. DuckDB is
designed for analytical query workloads (OLAP), where changes to the data are
expected to be large-scale
([DuckDB, "Why DuckDB", Fast](https://duckdb.org/why_duckdb)). This
application's workload is the opposite: many small transactional writes, one
post or comment at a time. DuckDB would not have prevented this kind of
defect on its own either: DuckDB v1.5.5 enforces foreign keys, but rejects
`ON DELETE CASCADE` when the table is created ("FOREIGN KEY constraints cannot
use CASCADE, SET NULL or SET DEFAULT", checked locally on 2026-11-23).

## Drawbacks, given the errors recognized today

- **Foreign keys are off unless each connection enables them.** Declaring
  them in the schema is not enough, and nothing fails when they are left off:
  the declaration reads as a guarantee that the engine does not apply. This
  is the defect above.
- **The choice was never recorded as a decision.** Its reasons lived in
  descriptive files that two later steps discarded, so the trade-off went
  unexamined until the defect surfaced.

## Why the defect is not fixed here

- The series corrects a defect at its origin and propagates the correction
  forward ([`ci-testing`'s ADR-000](https://github.com/CarpinetiOctavio/forum-app-ci-testing/blob/main/docs/decisions/ADR-000-resolving-forward-not-mirroring-backward.md)).
  The origin is `forum-app-ci-testing`, which is closed.
- Data integrity is outside TP7's scope (coverage, static analysis,
  end-to-end tests and quality gates).
- No guarantee this repository declares depends on it, so the exceeds-scope
  criterion (`docs/rules/documentation.md`, category d) does not apply. The
  real `DeleteComment` tests ([ADR-003](ADR-003-cypress.md)) check the
  authorship comparison, not what happens to related rows.

Alternatives rejected for the defect:

- **Enable foreign keys here** (`./database.db?_fk=1`, or
  `PRAGMA foreign_keys = ON` in `InitDB`). Rejected for the reasons above. The
  change is one line; what it lacks is a place in this repository's scope,
  not feasibility.
- **Delete a post's comments explicitly in `DeletePost`.** Rejected for the
  same reasons; it would also repeat in Go a rule the schema already declares.
- **Move to PostgreSQL.** Rejected for the same reasons, and for the cost
  recorded at the time.
- **Leave the documentation as it was.** Rejected: it described behavior the
  application does not have.

## Errors recognized in this record

- The reasons for choosing SQLite were lost in the 2026 rebuild of
  `forum-app-ci-testing` and in this repository's audit of `-legacy`, and were
  never included in `ci-testing`'s decision records.
- The audit took the cascade as confirmed from its declaration in the schema.
- Until 2026-11-23, `docs/SETUP.md` and `docs/diagrams/layered-architecture.svg`
  stated that the cascade works.

## Consequences

- Deleting a post leaves orphaned comment rows in `database.db`. Deleting a
  user by hand would leave their posts and comments; the application has no
  operation that deletes users.
- `docs/SETUP.md` describes the foreign keys as declared, not enforced, and
  links here. The README lists the defect under Limitations.
- A later repository that inherits this code decides on its own account
  whether to fix the defect or change the engine.
