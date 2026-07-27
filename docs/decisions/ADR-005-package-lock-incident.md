# ADR-005: package-lock.json desync incident (October 2025)

**Date:** 2026-07-04
**Status:** Resolved (documented retroactively)

## Context
This record is written months after the incident, from git history alone —
Octavio does not have an exact memory of the failure mode encountered at the
time. Per `docs/rules/documentation.md`, the cause below is framed as the most
probable explanation given the verifiable commit pattern, not as a certain
recollection.

The commit history around the comments feature, in order, is:

| Commit | Date | Message |
|---|---|---|
| `7ed7b94` | 2025-10-14 | Proyecto completo: backend + frontend + tests |
| `698d956` | 2025-10-14 | Update package-lock.json |
| `bcae814` | 2025-10-14 | Mod workflow CI/CD para backend y frontend |
| `938577a` | 2025-10-16 | Agregada funcionalidad de poder realizar comentarios, eliminarlos y mensaje de éxito al eliminar |
| `9de4ac8` | 2025-10-16 | Fix: sincronizar package-lock.json con dependencias |
| `09faaac` | 2025-10-16 | Resolución de dependencias en CommentList.tsx e reinstalación de npm |
| `f2829ec` | 2025-10-16 | reinstalación de dependencias |

Three consecutive commits on 2025-10-16 — all same day, immediately after the
comment feature commit (`938577a`) — deal explicitly with `package-lock.json`
synchronization and dependency reinstallation.

## Most probable cause (inferred, not remembered)
Adding the comment feature (`CommentList.tsx` and related components/tests)
most likely introduced or updated a frontend dependency (directly, or as a
transitive dependency pulled in by a new import) without regenerating
`package-lock.json` in the same commit. This would produce exactly the
symptom the next three commits describe fixing: `npm ci` in CI (which
requires an exact match between `package.json` and `package-lock.json`, unlike
`npm install`) failing or behaving inconsistently until the lockfile was
regenerated and committed.

This is inferred from the commit message pattern and the general mechanics of
`npm ci`, not from a saved CI log or exact memory of the failure output.

## Resolution
The lockfile was regenerated and committed (`9de4ac8`, `09faaac`, `f2829ec`),
restoring a `package.json`/`package-lock.json` pair that `npm ci` could
install from deterministically.

## Rationale for documenting this as an ADR
- **(c) Verified evidence from this repo's own history** — the commit pattern
  above is directly observable in `git log`, not asserted from memory.
- This incident is a concrete illustration of why `npm ci` (used in
  `.github/workflows/ci.yml`, not `npm install`) is the correct choice for
  CI: it is strict about lockfile/manifest consistency specifically so that a
  desynced lockfile fails fast and visibly, rather than silently installing a
  different dependency tree than what a developer tested locally.

## Alternatives considered
- **Switch CI to `npm install` instead of `npm ci`** to tolerate a desynced
  lockfile: rejected — this would hide the same class of problem instead of
  surfacing it, trading a visible, fixable CI failure for a silent
  dependency-tree drift between local and CI environments.
- **Leave the incident undocumented**, since it predates this documentation
  pass and is already resolved: rejected — the same class of failure (a new
  feature's dependency change without a lockfile update) can recur, and
  recording the pattern that revealed it is more useful than only recording
  the fix.

## Consequences
- No code change results from this ADR — it documents a past, already-fixed
  incident for the benefit of future contributors encountering the same
  `npm ci` failure mode.
- Confirms `npm ci` (not `npm install`) as the required command in any future
  CI step that installs frontend dependencies, precisely because it is the
  mechanism that surfaced this incident instead of masking it.
