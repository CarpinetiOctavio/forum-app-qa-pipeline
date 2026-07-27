# CLAUDE.md — forum-app-ci-testing

## Purpose of this file
Operating context for any AI assistant working in this repository. This repo is TP6
of a graded course series, now rebuilt as a standalone portfolio piece. It is the
first of three repos in a pipeline series (ci-testing → qa-pipeline → cloud-deploy),
each with a strictly bounded scope.

## Scope boundary — do not cross
This repo's scope is unit testing automation with GitHub Actions, limited to the
Services layer of a Go backend + React frontend forum app. Explicitly out of scope:
coverage gates, static analysis, integration tests, E2E tests, Docker, deployment.
These belong to later repos in the series and must not be introduced here, even if
they would technically improve the pipeline — doing so would break the deliberate
progression the portfolio series is built to demonstrate.

## Methodology (see ADR-000)
Decisions in this repo are not modeled on how forum-app-cloud-deploy solved the same
problem. Each decision must be grounded independently in software engineering
fundamentals of unit testing — established concepts and practice, not "the later repo
does it this way." Once a decision here is fully fundamented, it becomes the baseline
that propagates forward to qa-pipeline and then cloud-deploy — never the reverse.

## Initialization protocol
Before writing or modifying anything in a session:
1. Read every file in docs/rules/ in full.
2. Read every ADR in docs/decisions/ in full, in order.
3. Verify the current state of the repo against what the documentation claims (test
   counts, file structure, CI steps) — do not assume the docs are accurate.
4. Report findings and proposed next steps. Wait for explicit approval before writing
   anything.

## Decision-making authority
This assistant proposes and fundamenta options. It does not decide. Any change
affecting test behavior, scope, or documentation structure requires Octavio's explicit
approval before being written.

## Requirements for any proposed change, in order
1. Scope check — strictly within TP6's boundary?
2. Fundamentation check — grounded in a real software engineering concept about unit
   testing, not just "it works" or "the other repo does it this way"?
   A change that fails either check gets flagged, not implemented.

## Documentation standard
English for all prose (README, ADRs, SETUP, COMMANDS). Test names translated to
English (ADR-006) — unlike cloud-deploy's ADR-008, which keeps them in Spanish. The
two decisions don't conflict; they answer to different contexts (see ADR-000, ADR-006).

## AI usage disclosure
Claude acts as a conceptual auditor and writing assistant — never as decision-maker
for test design, mocking strategy, or scope boundaries. All design decisions were made
by Octavio Carpineti; Claude's role was surfacing inconsistencies, grounding proposals
in software engineering fundamentals, and drafting documentation for review.