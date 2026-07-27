# ADR-006: Test names translated to English

**Date:** 2026-07-04
**Status:** Accepted

## Context
The original test suite (backend Go function identifiers and frontend Jest
test descriptions) mixed English and Spanish naming — e.g.
`TestRegister_EmailVacio`, `test('muestra botón eliminar solo para posts
propios', ...)` — while the rest of the codebase's prose (comments, this
documentation) is being standardized to English. A decision was needed on
whether to translate these names, and — since a later, related repo in this
series (`forum-app-cloud-deploy`) faced the same question and answered it
differently in its own ADR-008 — why the two repos diverge without one
having copied the other's reasoning (see ADR-000).

## Decision
Every Spanish-named Go test function and every Spanish Jest test description
string is translated to precise, semantically equivalent English — not a
literal word-for-word translation — while leaving untouched: assertion lines
that check literal UI text or business-logic error strings still produced by
the (Spanish-language) application itself, and any mock data. See
`docs/decisions/ADR-000` for the rename list produced and its review process.

## Rationale
- **`docs/rules/testing.md`'s naming convention already requires this for any
  new test** (`Should_When`-style, English, precise about the business case).
  Renaming the inherited suite to the same standard removes the
  inconsistency of half the suite following a convention the other half
  predates.
- **The technical risk is real but was checked, not assumed away.** Several
  frontend test files interleave the test description string with
  assertions bound to literal, still-Spanish UI text (e.g.
  `screen.getByRole('heading', { name: /iniciar sesión/i })`,
  `screen.getByText('Eliminar')`). Every rename in this pass touched only the
  description-string argument of `test()`/`it()` — never an assertion line,
  never mock data, never the application's own UI text — and the full suite
  (23 backend + 34 frontend) was re-run after every file's renames to confirm
  nothing broke. Where a literal quote of the (Spanish) rendered UI text would
  have made an English test name misleading (e.g. a component still literally
  renders "No hay posts"), the new name describes the verified behavior
  instead of quoting text that would not match what actually renders (e.g.
  `shows an empty-state message when there are no posts`).

## Why this differs from cloud-deploy's ADR-008 — without adopting its reasoning
`forum-app-cloud-deploy`'s ADR-008 chose to leave its inherited Spanish test
names untouched, adding an English comment above each instead. That decision
is correct *for that repo*, under a constraint specific to it:
`cloud-deploy`'s own rules forbid modifying its inherited test suite unless a
test is provably broken, precisely because it is downstream of this series
and treats the suite as something received, not authored there. Renaming a
description string on the same line as a UI-coupled assertion, in a suite
that repo is committed not to touch, was correctly judged not worth the risk
for zero functional gain in that context.

This repo has no such inherited-suite constraint — it is the point of origin
for this test suite, not a downstream consumer of it. Per ADR-000, this
repo's decisions are grounded independently, in TP6's own scope and in
software-engineering practice around test naming, not in mirroring or
diverging from cloud-deploy for its own sake. The two ADRs reach different
conclusions because they answer different questions: cloud-deploy's is "should
I edit a suite I inherited and am committed not to touch," and this repo's is
"should the suite I originated follow its own stated naming convention." They
do not conflict; citing one to justify the other would be exactly the
mirroring ADR-000 rules out.

## Alternatives considered
- **Leave Spanish names as-is, matching cloud-deploy's ADR-008:** rejected —
  the constraint that makes that choice correct in cloud-deploy (an inherited
  suite under a no-touch rule) does not exist in this repo, which originates
  the suite and has already stated an English-naming convention for it in
  `docs/rules/testing.md`.
- **Add an English comment above each Spanish name without renaming:**
  rejected for the same reason — this repo is not under a constraint that
  makes avoiding the rename the safer choice; the actual rename is
  achievable at acceptably low risk once assertion lines and UI-text
  dependencies are checked individually, as done here.
- **Literal, word-for-word translation:** rejected — several Spanish names
  used constructions that read awkwardly translated literally (e.g.
  `NoEsAutor` → `NotTheAuthor` rather than `IsNotAuthor`) and instead re-
  expressed the exact business case each test verifies, per
  `docs/rules/testing.md`'s standing instruction to be "precise about the
  business case being verified — not a literal word-for-word translation."

## Consequences
- 16 Go test function identifiers and 20 Jest test description strings were
  renamed. Doc comments directly above renamed Go functions were updated to
  match, since a stale comment referencing an old function name is the exact
  documentation-honesty problem `docs/rules/testing.md` requires fixing, not
  leaving standing.
- Backend (23/23) and frontend (34/34) suites pass after the rename.
- Component `.tsx` source files and their rendered (Spanish) UI text were not
  touched — that is a product-copy decision out of this ADR's and this TP's
  scope.
- **The 23/23 and 34/34 counts above are a point-in-time verification
  record** — proof this rename didn't break anything at the time it was
  made — not a living count. Both totals have since grown (24 backend per
  ADR-008; 36 frontend, from tests added independently of both this ADR
  and ADR-008); see the README's Metrics table for the current counts.
