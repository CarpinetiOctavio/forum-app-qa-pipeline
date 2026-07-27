# ADR-009: Comment language and AAA-comment convention consistency

**Date:** 2026-07-21
**Status:** Accepted

## Context
ADR-006 translated every Spanish Go test function name and Jest test
description string to English. ADR-007 translated every backend error string
and every piece of rendered frontend UI text to English. Neither ADR's
declared scope covers a third surface that still had residual Spanish: internal
comments inside test files — the prose above `mock.On(...)` setup calls, and
the Arrange-Act-Assert section markers themselves. This surface was found
during the same code audit that produced ADR-008, while reading
`auth_service_test.go` and `post_service_test.go` line by line to fix the
tests affected by ADR-008's bcrypt change.

Grepping the full repository (`backend/internal`, `backend/cmd`,
`backend/tests`, `frontend/src`, `docs/*.md`) for Spanish-language patterns
confirms the residue is contained to three files:
`backend/tests/services/auth_service_test.go` (6 comment lines),
`backend/tests/services/post_service_test.go` (7 lines flagged — 5 comments to
translate, plus a 2-line pair that turned out to be an edit artifact rather
than documentation, see Decision 2 below), and `frontend/src/__mocks__/axios.ts`
(2 comment lines). Two of `post_service_test.go`'s lines were missed on the
initial pass and caught by a second full-repo grep before this ADR was
written. Spanish strings still present elsewhere in the repo — e.g. old test
names quoted inside ADR-006/007 as historical examples of what was renamed,
and Spanish git commit messages quoted verbatim in ADR-005's changelog table
— are citations of history, not untranslated code, and are correctly left
as-is.

Separately, the same three files were found to mix three different forms of
the Arrange-Act-Assert section comment: a bare `// ARRANGE` / `// ACT` /
`// ASSERT` (the majority pattern), the same marker followed by an inline
description already in English (e.g. `// ACT: User 1 deletes their own
post`), and the same marker followed by an inline description still in
Spanish (e.g. `// ARRANGE: Preparar el mock y datos de prueba`). Translating
the Spanish instances word-for-word without addressing this would still leave
the file internally inconsistent about what an AAA marker line looks like.

## Decision
1. All 13 Spanish comment lines across the three files listed above are
   translated to precise English — matching ADR-006's standing instruction to
   translate the business meaning, not word-for-word.
2. `post_service_test.go`'s `// ← AGREGAR ESTO` / `// ← FIN` pair (wrapping a
   block of user-setup code in `TestCreatePost_Success`) is deleted outright,
   not translated. It is not a documentation comment in any language — it is
   an edit marker left over from a prior change, and translating it would
   preserve noise instead of removing it.
3. Every `// ARRANGE`, `// ACT`, and `// ASSERT` marker across the three files
   is normalized to the bare form, matching this codebase's own existing
   majority convention. Where a marker previously carried a real, non-generic
   description (e.g. which user is performing the action), that description
   becomes its own comment line immediately before the marker, rather than
   being merged into it — the same shape `TestDeleteComment_NotTheAuthor`
   already used before this change (a plain descriptive comment, then a bare
   `// ACT` on its own line), which is the pattern the other instances are
   brought into line with, not a new pattern invented for this ADR.

Mock/fixture data representing user-authored content (e.g. `Content:
"Contenido"` in `post_service_test.go`) is deliberately left untouched, per
ADR-007's existing rule that such fixtures represent what a real user typed
into the app, not the project's own language surface.

## Why this is a separate ADR from ADR-008
ADR-008 documents security hardening: bcrypt, request-body size limits, and an
internal-error leak fix — each justified by a concrete exploitability
argument. This ADR documents comment language and formatting consistency —
justified by `docs/rules/documentation.md`'s general documentation-honesty
standard, not by security exploitability. Merging the two into one ADR would
dilute the "one ADR, one bounded decision" discipline ADR-006 and ADR-007
already established for this repo — a reader asking "why was this comment
translated" and a reader asking "why does this repo now hash passwords" are
asking about two unrelated axes of the same commit, and each deserves an ADR
that answers only its own question.

## Why this isn't already covered by ADR-006 or ADR-007
- ADR-006's decision is explicitly scoped to "every Spanish-named Go test
  function and every Spanish Jest test description string" — the string
  argument to `test()`/`it()`, or the Go function identifier itself. It does
  not mention comments.
- ADR-007's decision is explicitly scoped to the application's own runtime
  surface: backend error strings actually returned over HTTP, and frontend
  text actually rendered to a user. A comment inside a test file is neither —
  it is never returned by the API and never rendered in the UI.
- Internal test comments are consequently a third surface with no prior ADR,
  not a gap in either existing one.

## Alternatives considered
- **Fold this into ADR-008 as a subsection:** rejected, see "Why this is a
  separate ADR from ADR-008" above.
- **Translate the 13 lines and leave the AAA-marker inconsistency alone:**
  rejected — it would satisfy ADR-006/007's language standard by coincidence
  while leaving the actual inconsistency (three different shapes for the same
  marker) unaddressed, which is the same kind of documentation-honesty gap
  `docs/rules/testing.md` already requires fixing when found, not leaving
  standing.
- **Invent a new AAA-comment format for all three files:** rejected in favor
  of normalizing to the format already dominant in this codebase — introducing
  a fourth format would still leave three files each partially non-conformant
  to whatever the new standard was, for no benefit over adopting the format
  most of the suite already uses.

## Consequences
- No test behavior changed: this ADR is a comment- and formatting-only change,
  verified by running the full suite unchanged in outcome before and after
  (backend 24/24, frontend 36/36 — same counts as ADR-008 records, since both
  changes landed in the same review pass).
- Any future test added to these three files should use the bare
  `// ARRANGE` / `// ACT` / `// ASSERT` form, with any case-specific
  description as its own preceding comment line, not merged into the marker.