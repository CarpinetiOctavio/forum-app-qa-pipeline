# ADR-007: Application-wide language — English for all UI text and API error messages

**Date:** 2026-07-04
**Status:** Accepted

## Context
ADR-006 established English for test names and identifiers. Beyond that surface,
this repository still had a second, larger surface in Spanish: 24 error message
string literals returned by the backend API (`errors.New(...)` in `AuthService`
and `PostService`, plus `respondWithError`/`respondWithJSON` calls in the HTTP
handlers), and the rendered UI text of every React component (`Login`, `PostList`,
`CommentList`, `CreatePost`, `CommentForm`, `PostDetail`, `App`) — headings,
buttons, placeholders, confirm dialogs, and client-side error messages. A decision
was needed on whether this repository's language standard extends past developer-
facing artifacts (docs, comments, test names) into the application's own runtime
surface: what the API actually returns, and what a user actually sees running the
app locally per `docs/SETUP.md`.

## Decision
Every backend error string and every frontend UI string is in English. This
includes:
- All `errors.New(...)` messages in `AuthService` and `PostService`.
- The one error string in `SQLitePostRepository.DeleteComment`.
- All `respondWithError`/`respondWithJSON` message strings in `AuthHandler` and
  `PostHandler` (`invalid JSON`, `user not authenticated`, `invalid ID`, `post
  deleted`, etc.).
- All rendered text in every component: headings, button labels, placeholders,
  `window.confirm`/`alert` strings, and client-side fallback error messages.
- The two runtime log lines in `main.go`/`database.go` (`log.Println`,
  `log.Fatal`), since they are output a developer actually sees when following
  `docs/SETUP.md`'s "Running locally" steps.

Mock/fixture data in tests representing user-generated content — post titles,
post bodies, comment text (e.g. `'Mi primer post'`, `'Mi comentario'`) — is left
in Spanish. That is example content a user typed into the app, not the
application's own language surface, and translating it would misrepresent what
real user input looks like in a fixture meant to simulate it.

## Rationale
- **A single running application should present one coherent language, not a
  patchwork.** Prior to this change, the same request path could produce an
  English doc comment describing a function, next to a Spanish string that
  function actually returns over HTTP. That split is a consistency defect
  independent of any audience — a codebase whose narration (comments, ADRs,
  test names) is in one language while its actual observable behavior (API
  responses, rendered UI) is in another is describing itself inaccurately to
  anyone who runs it, not just to anyone who reads it.
- **This repository is being read and run by an audience that does not read
  Spanish.** This portfolio is prepared ahead of an international academic
  exchange; the people evaluating it will run `docs/SETUP.md`'s steps and
  interact with the running app directly, not only read the source. An error
  message or UI label is exactly as much a part of "the repository" as a
  comment is, for that audience — arguably more, since it is the one artifact a
  reader will actually see execute, without needing to read Go or TypeScript.
- **This is not a business-logic change**, and the evidence for that claim is
  checked, not asserted: every validation rule, authorization check, and
  control-flow branch in `AuthService`/`PostService`/`PostRepository` is
  byte-for-byte identical before and after this change — only the string
  literal communicating the outcome of each branch changed language. Backend
  coverage (54.1% of `internal/services`) and frontend coverage (50.24% of all
  files) are identical before and after, confirming no branch was added,
  removed, or altered — this ADR documents a translation, not a refactor.
- **Distinct from, and consistent with, ADR-006.** ADR-006 covers developer-
  facing identifiers (test names). This ADR covers the application's own
  runtime surface (API responses, rendered UI) — a different surface, but
  grounded in the same principle ADR-000 requires: this repository originates
  this codebase in the series and is not bound by any inherited "don't touch"
  constraint, so it is free to hold its own runtime output to the same English
  standard already applied to everything describing it.

## A discrepancy found while verifying against cloud-deploy, and how it was resolved
Per the verification instructions for this change, `forum-app-cloud-deploy`'s
equivalent files were read only to confirm each Spanish string in this repo had
a real, verifiable English counterpart to check semantic equivalence against —
never as the reason for making the change itself (see ADR-000). One discrepancy
surfaced during that check: `cloud-deploy`'s own `Login.test.tsx` translates
every other string in the file to English but leaves one mock/assertion pair —
`'Credenciales inválidas'` — in Spanish, even though the corresponding UI and
every other error case in that same file are in English. That looks like an
oversight in `cloud-deploy`, not a deliberate choice (nothing in that repo's ADRs
documents a reason to keep that one string Spanish while translating everything
around it). This repository does not mirror that inconsistency: the equivalent
string here was translated to `'Invalid credentials'`, resolved independently
against this ADR's own stated principle (one coherent language for the whole
running surface), not by copying `cloud-deploy`'s current state as-is.

## Alternatives considered
- **Translate comments, docs, and test names only; leave error messages and UI
  in Spanish:** rejected — produces exactly the split-language defect this ADR
  exists to close: an artifact that narrates itself in English while behaving
  in Spanish when actually run.
- **Introduce an i18n/localization framework** (`react-i18next`, `go-i18n`) to
  support both languages: rejected as disproportionate — this is a single-
  locale demonstration app for TP6; adding a localization layer to preserve a
  language the intended audience does not read would be scope creep with no
  TP6 requirement behind it.
- **Literal, word-for-word translation of every string:** rejected — several
  strings (e.g. UI copy with Spanish phrasing conventions like `"¿No tienes
  cuenta? Regístrate"`) were re-expressed idiomatically
  (`"Don't have an account? Sign Up"`) rather than translated mechanically,
  consistent with ADR-006's standing instruction to be precise about meaning,
  not literal about word order.

## Consequences
- 24 backend error strings (`AuthService`, `PostService`, `PostRepository`) and
  every corresponding test assertion in `auth_service_test.go` /
  `post_service_test.go` were updated together; the full backend suite
  (23/23) was re-run and confirmed passing after every file.
- All UI text in `Login`, `PostList`, `CommentList`, `CreatePost`,
  `CommentForm`, `PostDetail`, and `App` is in English; every corresponding
  assertion and mock response in `Login.test.tsx`, `PostList.test.tsx`,
  `CommentList.test.tsx`, and `authService.test.ts` was updated in the same
  step. The full frontend suite (34/34) was re-run and confirmed passing, and
  `npm run build` was confirmed to compile cleanly, after every component.
- Mock fixture data representing user-authored content (post/comment titles
  and bodies in `PostList.test.tsx` and `CommentList.test.tsx`) remains in
  Spanish, unchanged, on purpose — see Decision above.
- Coverage is unchanged: 54.1% backend, 50.24% frontend ("All files"), before
  and after — confirming this ADR records a translation, not a behavior change.
- **Every count and percentage in this section is a point-in-time
  verification record** — proof this translation pass didn't change
  behavior — not a living figure. Current state differs: 24/24 backend
  tests at 55.2% coverage following ADR-008's security hardening; 36/36
  frontend tests (grown independently of both this ADR and ADR-008) at
  52.19% ("All files"). Why the frontend percentage moved from 50.24% to
  52.19% between this ADR and now was not re-verified here — see the
  README's Metrics table for the current figures.
