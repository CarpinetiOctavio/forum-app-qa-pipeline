# ADR-001: Stack choice — Go + React over the course's .NET/Angular example

**Date:** 2026-07-04
**Status:** Accepted

## Context
TP6's assignment brief and in-class examples used .NET (backend) and Angular
(frontend) to demonstrate unit testing and mocking. The assignment itself does not
require a specific language or framework — it requires demonstrating unit testing
of business logic, a mocking strategy for external dependencies, and CI automation
of the test suite.

## Decision
Build the forum app in Go (backend) and React + TypeScript (frontend), using
testify (assert + mock) and Jest + React Testing Library respectively, instead of
the course's .NET/Angular example stack.

## Rationale
- **Testing concepts are language-agnostic.** The Arrange-Act-Assert pattern,
  dependency injection via interfaces, mock-based isolation of external
  dependencies, and the distinction between unit and integration testing are not
  specific to any language or framework — they are properties of the testing
  discipline itself. What TP6 evaluates is whether these concepts are understood
  and correctly applied, not fluency in a particular framework's syntax.
- **Deeper fluency in the chosen stack removes a confound.** Prior working
  fluency in Go and React means implementation time goes toward correctly
  designing tests and mocks, not toward learning unfamiliar framework syntax
  under time pressure — reducing the risk that a testing mistake is actually a
  language mistake.
- **Direct tool equivalence exists**, confirming the substitution changes no
  underlying capability required by the assignment:

  | Testing concern | .NET (course example) | Go + React (this repo) |
  |---|---|---|
  | Backend assertions | xUnit | testify/assert |
  | Backend mocking | Moq | testify/mock |
  | Frontend test runner | Jasmine/Karma | Jest |
  | Frontend HTTP mocking | Moq-equivalent HTTP fake | jest.mock('axios') |
  | CI/CD | GitHub Actions | GitHub Actions |

- **testify and Jest are each the low-friction default for their language,
  not an arbitrary pick.** testify is the standard assertion library for Go
  projects that don't need a full testing framework beyond the standard
  library's `testing` package — it adds `assert`/`mock` on top of what `go
  test` already provides, rather than replacing it. Jest ships preconfigured
  with Create React App with no additional setup, and already includes
  coverage collection out of the box — both were already the path of least
  resistance for this stack before any comparison to the course's .NET/Angular
  tooling was made.

## Alternatives considered
- **.NET + Angular, matching the course example exactly:** would have
  demonstrated the same testing concepts, but at the cost of implementation time
  spent on framework mechanics rather than test design, given no prior working
  fluency in that stack.
- **A stack mismatched between backend and frontend testing maturity** (e.g., a
  language with no mocking library as mature as testify or Jest): rejected,
  since TP6's core deliverable is mocking strategy — the chosen stack needed a
  first-class mocking library on both sides.

## Consequences
- Every testing concept demonstrated in this repo (AAA, mock-based isolation,
  dependency injection, coverage measurement) must be explainable independently
  of the specific stack — the stack choice does not by itself satisfy the
  assignment, the reasoning behind each test does.
- Tooling equivalence table above is the reference for translating this repo's
  choices back to the course's original example stack if ever required for
  grading comparison.
