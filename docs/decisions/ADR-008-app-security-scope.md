# ADR-008: Application security hardening applied post-v1.0.0, and the scope boundary around it

**Date:** 2026-07-21
**Status:** Accepted

## Context
This repo was tagged `v1.0.0` and treated as closed. While auditing
`forum-app-qa-pipeline` — the next repo in this series, built as a copy of this
repo's `v1.0.0` — two problems were found in the app's own source code, not in
qa-pipeline's testing/CI additions: passwords stored and compared in plain text,
and no upper bound on the size of any JSON request body. Both exist identically
in this repo, because qa-pipeline's backend and frontend are a direct copy of
this one. Per ADR-000's methodology (corrections propagate forward, never
backward), these are fixed here, not in qa-pipeline, so that qa-pipeline's next
copy of this codebase inherits the fix instead of re-introducing the same
problem one repo later.

A broader audit of the app's own design (not just these two points) was run
against this repo's code directly, independent of qa-pipeline's audit, to
confirm which findings were shared code and which were specific to one mirror.
That audit is the basis for the decisions below and for the scope boundary this
ADR draws around what gets corrected now versus documented as an accepted,
out-of-scope limitation.

## Decision
Three corrections are implemented in this repo, each scoped tightly to the
finding it addresses:

1. **Passwords are hashed with bcrypt.** `AuthService.Register` hashes the
   incoming password with `bcrypt.GenerateFromPassword` before it reaches the
   repository; `AuthService.Login` compares with
   `bcrypt.CompareHashAndPassword` instead of a direct string comparison. The
   external error contract is unchanged — `Login` still returns
   `"invalid credentials"` on any mismatch, whether the email doesn't exist or
   the password doesn't match, exactly as before.

2. **Request bodies are capped at 1MB.** `Register`, `Login`, `CreatePost`, and
   `CreateComment` — the four handlers that decode a JSON body — wrap
   `r.Body` in `http.MaxBytesReader(w, r.Body, 1<<20)` before decoding, so an
   oversized body is rejected instead of fully buffered and parsed.

3. **`GetAllPosts` no longer returns a raw internal error to the client.** Of
   the 9 call sites in the handlers layer that respond with `err.Error()`, 8
   propagate a message the Services layer already constructs deliberately
   (`"post not found"`, `"invalid credentials"`, `"title is required"`, etc.) —
   that is intentional API design, not a leak, and those exact strings are
   already asserted by the existing test suite. `GetAllPosts`'s `500` path was
   the one exception: it returned whatever `postRepo.FindAll()` produced,
   unfiltered. It now logs the real error server-side (`log.Println`) and
   responds with a generic `"internal server error"` message.

### Password length ceiling: explicit validation, not a propagated bcrypt error
`bcrypt.GenerateFromPassword` itself rejects any input over 72 bytes.
`Register` already validated a 6-character minimum but had no maximum, so a
password over the limit would have reached bcrypt and surfaced whatever error
`bcrypt` itself produces.

**Alternative considered and rejected:** let that bcrypt error propagate to the
client as-is. Rejected because the handler layer responds with `err.Error()`
for `Register`'s `400` path (see point 3 above on why that's normally correct,
curated behavior) — an unredacted, library-authored error string reaching the
client through that same path would reintroduce, in the same commit that
fixes point 3's `GetAllPosts` leak, the same category of problem: a message
this project did not write or review, reaching the client verbatim.

**Decision:** an explicit check, `len(req.Password) > 72`, returns a named,
project-authored error — `"password must not exceed 72 characters"` — before
bcrypt is ever called. This is consistent with how every other validation rule
in `Register` already works: a named business rule with its own test
(`TestRegister_PasswordTooLong`), not a library's internal error surfacing
through the API.

The same 72-byte limit was replicated in the frontend (`maxLength={72}` on
`Login.tsx`'s password `<input>`) for UX consistency, once the backend limit
was decided — this is not a separate decision, it's the same 72-byte ceiling
applied at the point a user can act on it before submitting, rather than only
after the backend rejects the request. Not covered by its own unit test: per
ADR-002's criterion, a unit test is warranted for a branch that discriminates
a business-rule or safety-relevant outcome, not for the mere presence of a
static attribute — `maxLength` doesn't branch on anything, and no other
static attribute on this same input (`type="password"`, `required`) has a
dedicated test either, so adding one only for this attribute would be the
inconsistent choice, not the safe one.

### MaxBytesReader limit: 1MB uniform across all 4 handlers, not tuned per endpoint
`Register`/`Login` receive only email, password, and username — a few dozen
bytes in practice. `CreatePost`/`CreateComment` could reasonably justify a
different (likely smaller, given this is a text-only forum with no
attachments) limit than the auth endpoints, tuned specifically to expected
content length.

**Alternative considered and rejected:** four endpoint-specific limits.
Rejected for lack of grounding — assigning a distinct number to each endpoint
without real production traffic data to size against would replace one
unmeasured placeholder with four unmeasured placeholders, which is not an
improvement.

**Decision:** a single `1<<20` (1MB) constant applied uniformly. This is
recorded explicitly as a placeholder chosen for being generous enough not to
reject any legitimate forum post while still bounding the worst case, not as a
number derived from measured usage. Revisit if real traffic data becomes
available.

### 413, not 400, for an oversized body
Go exposes `http.MaxBytesError` (a typed error, not just a string) as of Go
1.19; this module already targets Go 1.24.1. After `Decode` fails, the handler
checks `errors.As(err, &maxBytesErr)`: if it matches, it responds
`http.StatusRequestEntityTooLarge` (413) with `"request body too large"`; if
not, it falls through to the existing `http.StatusBadRequest` (400),
`"invalid JSON"` path.

**Alternative considered and rejected:** keep responding `400` for both cases,
since `Decode` fails either way. Rejected because a body that is too large is
not malformed JSON — collapsing both into `"invalid JSON"` would tell a client
the wrong thing about what to fix. `errors.As` on the typed error is the
idiomatic Go mechanism for this exact distinction; it does not require
parsing the error string or trusting `Content-Length`, which a client can
omit or misreport.

No existing test asserts `"invalid JSON"` for an oversized-body request — the
handlers layer is out of unit-test scope in this repo (ADR-002), and no test
exercised this path before this change, so there was nothing to break.

## Known limitations — accepted, not corrected here
These were found during the same audit and are deliberately left as-is in this
repo. Each is a real weakness in a production context; each is judged
acceptable for an academic CI/CD portfolio demonstration, for the reason
stated:

- **No real authentication.** `X-User-ID` is a client-supplied header with no
  token, session, or signature behind it — any request can claim to be any
  user. Fixing this would mean introducing JWT or server-side sessions, which
  is a scope-level architectural change to how every handler authenticates a
  caller, not a contained fix like the three above. It is out of this repo's
  bounded scope (unit testing automation, per `CLAUDE.md`) and would need its
  own ADR if ever undertaken, in whichever repo in the series is judged the
  right place for it.
- **CORS allows any origin** (`Access-Control-Allow-Origin: *` in
  `router.go`). This is a direct consequence of the point above — CORS
  restriction has no real teeth without something to authenticate in the
  first place — and is accepted for the same reason: fixing it meaningfully
  requires fixing authentication first.
- **No pagination on `GetAllPosts`/`GetComments`.** Both return every row with
  no `LIMIT`. Acceptable at the dataset size a course demo ever reaches;
  revisiting this is a product/API-design decision, not a security patch of
  the kind this ADR scopes.
- **Dependency vulnerabilities**: `npm audit` on the frontend reports 55
  (2 critical), all transitive within Create React App's dev-only tooling
  (`react-scripts`/`webpack-dev-server`), not in the production bundle.
  `govulncheck` on the backend reports 18 reachable vulnerabilities, all in
  the Go standard library itself, tied to the local/CI Go toolchain's patch
  version rather than to this project's code or its three direct
  dependencies (`gorilla/mux`, `go-sqlite3`, `testify`, all clean). Neither is
  a defect in this project's own code.

## Consequences
- New dependency: `golang.org/x/crypto v0.48.0` (pinned below the `v0.49.0+`
  releases that require Go 1.25, to stay on this module's `go 1.24.1` and
  match the CI workflow's pinned `go-version: '1.24'`).
- `AuthService.Register`'s validation sequence gained one new named error path
  (`"password must not exceed 72 characters"`), with its own test.
- `TestLogin_Success` and `TestLogin_IncorrectPassword` now generate a real
  bcrypt hash in `ARRANGE`, using `bcrypt.MinCost` rather than `DefaultCost` —
  the property under test here is the comparison logic, not production-grade
  hashing cost, so there is no reason to pay bcrypt's real cost twice per
  test run. (See the note below on `TestRegister_Success` for the one test
  in this suite that cannot make this same choice, and why.)
- `TestRegister_Success` gained an assertion that the password reaching
  `UserRepository.Create` is bcrypt-hashed, never the plaintext input
  (captured via testify's `mock.Arguments.Run`, then verified with
  `bcrypt.CompareHashAndPassword`). Unlike the two Login tests above, this
  one cannot use `bcrypt.MinCost`: it calls the real `AuthService.Register`,
  which hashes with `bcrypt.DefaultCost` internally — the cost is dictated by
  production code, not a test-side choice, so this is the one test in the
  suite that pays bcrypt's real ~60-100ms cost on every run. That's expected
  for this single test, not a regression to chase — mocking the hashing
  itself to avoid it would test nothing.
- Backend suite: 24/24 passing (23 before this change, +1 new test). Backend
  coverage of `internal/services`: 55.2% (54.1% before, per ADR-007's
  baseline) — the increase reflects the one new validated branch, not a
  behavior change elsewhere.
- Frontend: `Login.tsx`'s password input gained `maxLength={72}` (see the
  password length ceiling section above for the UX-consistency reasoning) —
  the only frontend source change in this ADR. Suite still 36/36 passing;
  `npm run build` still compiles cleanly. No new frontend test, per the
  reasoning above.
- ADR-004, ADR-006, and ADR-007 each cite backend/frontend test counts and
  coverage percentages (23/23, 34/34, 54.1%, 50.24%) that were accurate when
  those ADRs were written and verified. Those citations are point-in-time
  verification records — proof that a specific change of theirs didn't break
  anything at the time — and are intentionally left as-is rather than
  updated: doing so would misrepresent what was actually verified when each
  of those ADRs was accepted. This ADR's changes are why the backend figure
  now differs (24/24, 55.2%); the frontend figure differs for an unrelated
  reason (36/36 tests, grown independently of this ADR, sometime between
  ADR-007 and this one). The README's Metrics table is the current,
  living source of truth for both.
- This is the concrete instance ADR-000's forward-propagation methodology
  describes, not an exception to it: found via qa-pipeline's audit, corrected
  here first, so that qa-pipeline's copy of this codebase — and
  cloud-deploy's after it — inherits the fix rather than re-deriving it. This
  ADR is what advances this repo's tag from `v1.0.0` to `v1.1.0`: the TP6
  scope that `v1.0.0` closed is not being reopened or invalidated, it is
  being extended with a post-close finding that the series' own methodology
  requires to be resolved at its origin point.