# ADR-003: Mocking strategy — mock the boundary, not the logic

**Date:** 2026-07-04
**Status:** Accepted

## Context
`AuthService` and `PostService` depend on `UserRepository`/`PostRepository`
(SQL access) to do their work; `authService` and `postService` on the frontend
depend on `axios` to reach the backend over HTTP. A decision was needed on
which of these dependencies get replaced with a test double, and which get
exercised for real during a unit test run.

## Decision
Mock exactly the I/O boundary of the unit under test, and nothing else:
- **Backend:** `UserRepository` and `PostRepository` are mocked via
  `testify/mock`, injected into the real `AuthService`/`PostService` through
  their constructors. The service's own logic (validation, authorization,
  error propagation) runs unmocked.
- **Frontend:** `axios` is mocked via `jest.mock('axios')` and a manual
  `__mocks__/axios.ts`. `authService`/`postService`'s own logic (building the
  request, unwrapping `response.data`) runs unmocked.

## Rationale
- **Isolation is the defining property of a unit test.** A unit test verifies
  the behavior of one unit by removing its dependency on collaborators that
  are slow, non-deterministic, or require external state (a running database,
  a live network call) — replacing them with a test double under the test's
  control. Mocking anything closer to the unit under test than its I/O
  boundary (e.g., mocking `AuthService` itself while testing `AuthService`)
  would make the test tautological; mocking anything further out (e.g.,
  running a real SQLite instance) would make it an integration test, which is
  a different — and, for this repo, out-of-scope — concern (see ADR-002).
- **Both mocking points sit exactly at that boundary already, by
  construction.** The backend's `Repository` is defined as a Go interface
  specifically so a concrete implementation (SQLite) can be swapped for a
  test double via dependency injection, without changing `Service` code. The
  frontend's `authService`/`postService` call `axios` directly as their only
  external dependency; mocking the `axios` module is the narrowest point that
  isolates them from the network.
- **This was already the established pattern in this repo's own history**,
  not a change introduced now: `testify` is declared in `backend/go.mod`
  specifically for this purpose, and `frontend/src/__mocks__/axios.ts`
  predates this documentation pass. This ADR records and justifies an
  existing choice, rather than introducing a new one.

## Alternatives considered
- **No mocking — exercise the real repository/HTTP call:** rejected. Tests
  would require a running database and/or backend server, making them slow,
  order-dependent, and liable to fail for reasons unrelated to the business
  logic under test (see `docs/rules/testing.md`: "no test hits a real
  database, HTTP call, or filesystem").
- **Mock the Service/component itself:** rejected — a test that mocks the
  unit it claims to test verifies nothing; it only confirms the mock returns
  what it was configured to return, which `docs/rules/testing.md` explicitly
  names as an invalid test for this scope.
- **Mock at a coarser boundary** (e.g., mock the entire HTTP client
  configuration rather than the `axios` module itself): considered
  unnecessary — `axios` is already the narrowest seam available without
  introducing an additional abstraction layer purely to ease testing, which
  would be scope creep for TP6.

## Consequences
- Tests are fast (no I/O) and deterministic (mock responses are fixed per
  test case), and can simulate error conditions (a duplicated email, a
  network failure) that would be hard or slow to reproduce against a real
  database or server.
- These tests verify that `Service`/`authService`/`postService` behave
  correctly given a certain repository/HTTP response — they do not verify
  that the real SQL queries are correct, or that the real HTTP contract with
  the backend matches what the mock assumes. That verification is an
  integration/E2E concern, out of this repo's scope.
