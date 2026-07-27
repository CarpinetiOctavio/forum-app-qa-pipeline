# Operating rules — testing (forum-app-ci-testing)

## Scope of this repo's tests
Unit tests only, Services layer only (AuthService, PostService in backend;
authService, postService in frontend). Repository and Handlers layers are
deliberately untested at the unit level here — that's a decision (ADR-002), not an
omission, and it doesn't get "fixed" in this repo.

## What counts as a valid test for this scope
- Follows AAA (Arrange-Act-Assert).
- Mocks every external dependency of the unit under test — no test hits a real
  database, HTTP call, or filesystem.
- Tests behavior, not implementation — a test that verifies a mock returns what it
  was configured to return is not equivalent to a test verifying the real business
  rule.
- Test names in English (ADR-006), precise about the business case being verified —
  not a literal word-for-word translation from Spanish.

## Before writing or modifying any test
1. Confirm the change is strictly about unit testing of the Services layer — not
   coverage thresholds, integration behavior, or anything belonging to qa-pipeline.
2. Ground the change in an actual software engineering concept about unit testing
   (isolation, determinism, mock vs. stub distinction) — name the concept in the
   accompanying ADR, not just "this makes the test better."
3. If a test's label or comment doesn't match what the test actually verifies, that's
   a documentation-honesty problem to flag and fix, not to leave standing.

## What NOT to do
- No coverage gates, static analysis, or integration/E2E tests here.
- Do not resolve an inconsistency by copying how forum-app-cloud-deploy resolved the
  analogous case. Resolve it independently, grounded in this repo's own scope and
  literature.
- Do not silently "improve" test behavior without flagging it first.