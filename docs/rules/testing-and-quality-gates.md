# Operating rules — testing and quality gates (forum-app-qa-pipeline)

## Coverage scope
The 70% gate applies to `internal/services/` (backend) and the components/services
layer (frontend), per `ADR-001-coverage-strategy`. Handlers, Repository, Router,
Database, Models are deliberately excluded — that is a documented decision, not a
gap. Don't propose adding coverage requirements to an excluded layer without a new
ADR justifying why the original exclusion no longer holds.

## Branch protection ruleset must track new pipeline jobs
The ruleset on `main`/`staging` was imported from `ci-testing` and, as of this
repo's starter setup, has exactly one required status check: `Test Summary` —
the job name inherited from `ci-testing`'s own pipeline. It does not
automatically pick up new jobs added later.

When the SonarCloud and Cypress jobs are added to `ci.yml`, they must also be
added by hand as required status checks (Settings → Rules → edit the ruleset →
add each new job's exact name under "Require status checks to pass"). Skipping
this step means the quality gates TP7 declares (block the pipeline if SonarCloud
or Cypress fails) exist in the workflow file but don't actually block a merge —
the ruleset would still only be watching `Test Summary`. Treat "add a new
required job" as part of implementing that job, not a follow-up task to
remember separately.

## Before changing `sonar.coverage.exclusions` or any SonarCloud config
An exclusion added to make the Quality Gate pass is not a valid reason on its own —
it needs the same justification bar as an ADR (why this file/layer shouldn't be
measured, not just that measuring it is inconvenient). If a gate fails because
excluded/untested code shows up as uncovered, the fix is understanding why it's
failing, not silently widening the exclusion list.

## What a test actually proves — state it explicitly
Distinguish, in both code review and documentation, between what a test verifies
and what it appears to verify. Concretely: a mocked repository test proves the
service's logic is correct given an input — it does not prove the input itself is
trustworthy. A Cypress test using `cy.intercept()` proves UI behavior given a
response — it does not prove the backend produces that response, or that an
unauthenticated request would be rejected. When writing or reviewing a test,
name the actual guarantee, not the apparent one.

## Recreating scenarios from `forum-app-qa-pipeline-legacy`
Test scenarios catalogued in `docs/audits/qa-pipeline-legacy-transferable-knowledge-results.md`
as adding real coverage are a reference for *what* to test, not *how*. Write the
test code from scratch against this repo's actual implementation — don't adapt
`-legacy`'s test code, which was never validated against a documented testing
standard of its own (see that repo's own audit for why).

## Authentication vs. authorization
This repo inherits `forum-app-ci-testing`'s authentication model as-is (see that
repo's `ADR-008` for the known limitation and its rationale for deferral). Tests
written here that touch authorization logic (e.g., "only the author can delete")
must document that they validate the authorization comparison, not identity
verification — the latter doesn't exist yet in this codebase. Don't write or
review such a test as if it closes that gap.