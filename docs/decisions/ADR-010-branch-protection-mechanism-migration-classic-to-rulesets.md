# ADR-010: Branch protection mechanism — migrating from classic rules to rulesets

**Date:** 2026-07-24
**Status:** Accepted

## Context
ADR-004 established the `feature/* → staging (PR) → main (PR)` branching
model and noted that the model's enforcement — the actual thing stopping a
direct push to `staging` or `main` — depends on a branch protection rule
configured on GitHub itself, not on anything `ci.yml` or Git alone can
guarantee. At the time ADR-004 was written, that protection was configured
using GitHub's classic branch protection rules, the long-standing mechanism
for this.

The question of which mechanism to use resurfaced while retaking the
evidence screenshots for this protection (`docs/screenshots/
04-branch-protection-blocked-staging.png` and `05-branch-protection-blocked-
main.png`): the first attempt captured Git's own non-fast-forward rejection
(a local, mechanism-independent error caused by pushing from a stale branch
tip), not GitHub actually blocking a legitimate push — so the screenshots
needed to be retaken regardless of which mechanism was configured underneath.
That reshoot is what forced the question: reconfigure the same classic rule
and reshoot, or migrate to GitHub's newer ruleset mechanism first and shoot
the evidence against that instead.

GitHub has been moving deliberately away from classic branch protection rules
toward rulesets: a sibling legacy mechanism, tag protection rules, has
already been deprecated outright in favor of rulesets on GitHub Enterprise
Server, and GitHub's own documentation frames rulesets as having structural
advantages classic rules do not — multiple rulesets evaluating simultaneously
without silently overriding each other, an `evaluate` mode that reports what
a rule would do before it enforces anything (available on paid Enterprise
plans, not on this repository's plan — see "Migration approach" below), and a first-class REST/GraphQL
representation intended for managing repository settings as code. Classic
rules are not deprecated as of this writing, but they are not receiving that
investment either.

## Decision
Migrate `staging` and `main`'s branch protection from classic rules to a
ruleset.

### What was actually debated
Two candidate justifications for migrating were considered and rejected
before landing on the one actually used:

**Rejected justification: "rulesets are becoming the industry standard, worth
demonstrating familiarity with them."** This was the original prompt for
reconsidering the mechanism, raised in the context of a different portfolio
repo. It is rejected as the stated reason here because it optimizes for what
the choice signals to a reader, not for what problem it solves in this repo
— exactly the axis this portfolio's own stated criterion rules out (the
project's own guiding principle is that the central criterion is
demonstrating an applied reasoning pattern, not breadth of stack or tooling). An ADR whose real reason is "so it looks current" does
not survive being asked "what problem did classic actually cause you?" — the
honest answer would be none.

**Rejected justification: "classic already satisfies what this repo actually
needs, so keep it."** This was the first counter-argument considered.
Rulesets' headline advantages — layering multiple rules, organization-wide
reuse across many repositories — are genuinely irrelevant to a single-branch,
single-committer repository; there is no fleet of repos here to standardize
across, and no overlapping rules to reconcile. On those specific advantages,
classic is not disadvantaged for this repo's actual scale. But "it already
works" is not, by itself, a reason to keep something — it is only a reason if
switching costs something classic-remaining does not. Refusing to switch
purely because the current tool functions, with no argument for why
*continuing* to use it is better than the alternative, is inertia dressed as
a principled position — it fails the same test as adopting rulesets for
optics would have: neither is grounded in an actual cost or benefit specific
to this repo.

**Actual justification: no compensating benefit to staying, and equal cost to
switching.** The evidence screenshots had to be reconfigured and retaken
either way, because the first attempt captured the wrong kind of rejection —
so migrating to a ruleset costs nothing beyond what reconfiguring classic
correctly would have cost. Given equal cost, the deciding factor is which
mechanism GitHub is actively investing in: rulesets are, classic is not (and
GitHub has already fully deprecated at least one sibling classic-era
mechanism — tag protection rules — in favor of rulesets). "It already works"
is not a compensating benefit against a mechanism the platform itself is
steadily deprioritizing, when the alternative costs the same to configure
right now, while it's a small, low-stakes change, rather than later, as a
forced migration under time pressure.

## Migration approach
GitHub's `Evaluate` enforcement status — which would report what a ruleset
would block without actually enforcing it — is not available on this
repository's plan: it is documented only for GitHub Enterprise Cloud and
GitHub Enterprise Server, and does not appear as an option when creating a
ruleset on GitHub Free, Pro, or Team. The ruleset here is created directly in
`Active` status.

In place of `Evaluate`'s non-blocking observation, the ruleset's behavior is
validated the same way the previous classic rule's rejection was verified:
an empty, disposable commit (`git commit --allow-empty`) is pushed directly
to `staging` and to `main` immediately after the ruleset is created, and the
push is expected to be rejected by GitHub. This is a materially different
kind of test than `Evaluate` mode provides — it is a live, blocking attempt
rather than a passive observation — but it is a reasonable substitute here
specifically because this repository has a single committer and no ongoing
collaborator activity that a live test could disrupt; `Evaluate` mode exists
primarily to avoid inconveniencing real contributors during rollout, a
concern that does not apply to a repository at this scale.

## Repository plan pre-requisite
Rulesets themselves are available on GitHub Free for public repositories,
which this repository is — only the `Evaluate` enforcement status
specifically requires a paid Enterprise plan. The migration below does not
depend on any plan upgrade.

## Relationship to ADR-004
This does not reopen, amend, or contradict ADR-004. The branching model ADR-
004 established — `feature/* → staging (PR) → main (PR)`, no `develop`
branch — is unchanged; only the GitHub mechanism enforcing that model
changes. ADR-004's record that classic protection was configured and
verified at the time it was written is left as-is, as an accurate,
point-in-time account of what was true then — the same reasoning ADR-008
already applied to not rewriting ADR-004/006/007's historical test-count
citations. This ADR is the subsequent decision that supersedes the
*mechanism* alone.

## Consequences
- `docs/screenshots/04-branch-protection-blocked-staging.png` and
  `05-branch-protection-blocked-main.png` are replaced with captures taken
  against the ruleset, showing GitHub's actual protected-branch rejection
  (not the earlier non-fast-forward Git error, and not a screenshot of
  classic's equivalent rejection either).
- No change to `.github/workflows/ci.yml`, no change to any application or
  test code — this is a repository-settings decision only.
- `main`/`staging`'s protection is now configured as a ruleset rather than a
  classic branch protection rule. The two conditions ADR-004 originally
  specified (PR required, CI passing check required) carry over unchanged.
  Two conditions are added beyond ADR-004's original two: `Block force
  pushes` and `Restrict deletions`, both protecting the same asset ADR-004
  already establishes as the reason `main` exists — its tagged, stable
  history — against accidental rewriting or accidental deletion of
  `staging`/`main`, rather than against anything ADR-004 didn't already
  care about.