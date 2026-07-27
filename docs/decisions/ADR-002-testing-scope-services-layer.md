# ADR-002: Testing scope limited to the Services layer

**Date:** 2026-07-04
**Status:** Accepted

## Context
The application has four backend layers (Handlers, Services, Repository,
Database) and a frontend split between HTTP services (authService,
postService) and React components (Login, PostList, CommentList, CreatePost,
CommentForm, PostDetail). TP6 requires demonstrating unit testing and mocking,
not full coverage of every layer and component. A decision was needed on which
units justify a unit test within this TP's scope, and which do not.

## Decision
Unit tests are written only for:
- **Backend:** the Services layer (`AuthService`, `PostService`) — Repository
  and Handlers are not unit-tested here.
- **Frontend:** the HTTP services (`authService`, `postService`) and the
  components whose behavior branches on a business rule with more than one
  materially different outcome — `Login` (authentication success vs. failure),
  `PostList` and `CommentList` (delete button visibility gated by author
  identity). `CreatePost`, `CommentForm`, and `PostDetail` are not unit-tested
  here.

## Rationale
- **A unit test's value is proportional to the number of behaviorally distinct
  branches it can discriminate.** `AuthService` and `PostService` contain every
  validation rule and authorization check in the system (email format, password
  length, "only the author may delete" for both posts and comments) — this is
  where a regression would silently break a business rule. `Repository` methods
  execute SQL with no branching logic of their own; their correctness is a
  question of query correctness against a real database, which is an
  integration-testing concern, not a unit-testing one. `Handlers` decode HTTP,
  extract a header, and delegate to a service — a thin translation layer with
  no business rule to discriminate.
- **The same reasoning applies to component selection on the frontend.**
  `Login`, `PostList`, and `CommentList` each have at least one conditional
  outcome tied to a business rule (auth success/failure; permission-gated
  delete visibility). `CreatePost`, `CommentForm`, and `PostDetail` do hold
  local component state (`useState` for loading/error, and in `PostDetail`'s
  case conditional rendering on fetch status) — but none of that state
  branches on a business rule with more than one materially different
  system-level outcome to verify; it is generic request-lifecycle UI state,
  not evidence of an untested authorization or validation path. This is the
  operative distinction, not "these components have no logic," which several
  earlier internal drafts (now removed) incorrectly claimed.
- **This is an explicit, bounded scope decision of TP6**, not an emergent
  outcome of running out of time. TP6's deliverable is demonstrating the
  unit-testing and mocking discipline on the highest-risk layer; a full test
  pyramid — Repository correctness via integration tests, Handler behavior via
  E2E tests, remaining components via additional unit tests — is the declared
  scope of `forum-app-qa-pipeline`, the next repo in this series.

## Alternatives considered
- **Unit-test every layer and component:** would require either testing SQL
  execution in isolation (not meaningful without a real or in-memory database,
  which is an integration test by definition) or writing component tests with
  no business-rule branch to assert on, adding test count without adding
  discriminating power.
- **Unit-test Repository against an in-memory SQLite database:** rejected for
  this repo specifically because it blurs the unit/integration boundary this
  TP is meant to teach — a test that talks to any real database, in-memory or
  not, is an integration test by the same definition used to justify mocking
  Repository in the first place (see ADR-003).

## Consequences
- `Repository`, `Handlers`, `CreatePost`, `CommentForm`, and `PostDetail`
  remain without unit-level tests in this repo. That is a documented, bounded
  decision — not a gap to be silently closed here. Closing it, if warranted,
  belongs to `forum-app-qa-pipeline`.
- Any reviewer asking "why isn't X tested" can be answered by this ADR without
  needing to infer intent from coverage numbers alone.

## Post-documentation audit finding

While reviewing the frontend coverage output during this repository's
documentation pass, two uncovered branches were found in components already
inside this ADR's declared scope (`Login` and `PostList`) that qualify as
behavior with real consequence under the criterion this ADR already states —
not new scope, a gap inside scope the original development did not catch:

- **`Login.tsx`, the registration branch** (`authService.register(...)` +
  `onLoginSuccess(user)`): this branch determines whether completing
  registration actually leaves the user logged in. It was untested — the
  existing suite covered the login path and the toggle-to-register UI
  transition, but never submitted the registration form itself. If this
  branch silently stopped calling `onLoginSuccess`, or called the wrong
  service method, a user could complete registration and never be logged in
  — a real behavioral bug, not a cosmetic one.
- **`PostList.tsx`, the `window.confirm` guard** (`if (!window.confirm(...))
  return;` before calling `deletePost`): this branch is the only thing
  standing between a cancelled confirmation dialog and an irreversible
  delete. It was untested — the existing suite covered the case where the
  user confirms, never the case where they cancel. If this guard were
  inverted or removed, clicking "Cancel" would still delete the post.

Both gaps were closed: `Login.test.tsx` gained a test asserting that
submitting the registration form calls `authService.register` with the
correct credentials and calls `onLoginSuccess` with the returned user;
`PostList.test.tsx` gained a test asserting that when `window.confirm`
returns `false`, `deletePost` is never called. Both are verified against the
mocking boundary already established for these files (see ADR-003), not
against new infrastructure.

The bar for adding these two tests is the same bar this ADR already
declared for what belongs in scope: a branch that discriminates a
business-rule or safety-relevant outcome, not the mere presence of a
conditional. This is not scope expansion — it is closing a hole inside the
scope this ADR already draws, found by reading the coverage report rather
than assuming its gaps were all accounted for.

The remaining uncovered lines in the same coverage report — `App.tsx`,
`index.tsx`, `reportWebVitals.ts`, `CreatePost.tsx`, `CommentForm.tsx`,
`PostDetail.tsx`, and the generic error-handling `catch` blocks in
`CommentList.tsx` and `PostList.tsx` (`console.error`/`alert` calls with no
conditional business-rule branch of their own) — were audited in this same
review and classified as correctly out of scope under this ADR's own
criterion: none of them branch on a business rule with a verifiable,
real-world consequence at the unit level. Their 0% or partial coverage is
the correct outcome of that criterion, not an outstanding gap. A reader of
the coverage report — including someone with no prior context, such as a
recruiter — should read every red line in that output against this
criterion: does this specific line branch on a rule that could silently
break something real, or is it request-lifecycle plumbing (loading state,
a fixed error string, a null-check on an optional callback)? The two lines
closed above answered "yes." Every other uncovered line in this repository
answers "no," and is documented here as a deliberate, audited scope
boundary — not an omission.
