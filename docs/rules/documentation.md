# Operating rules — documentation (forum-app-ci-testing)

## ADR requirements
Every ADR must include: the problem/question, the alternatives actually considered,
the decision, and the justification. The justification must trace to one of: (a) a
concept from software engineering literature/practice on unit testing, (b) an
explicit scope boundary of TP6, or (c) verified evidence from this repo's own history.
"Because the other repo does it this way" is not a valid justification on its own.

## Before writing an ADR
Confirm with Octavio any fact that can't be verified from code or git history directly
(why an incident happened, whether a choice was deliberate or a shortcut). Frame
inferred causes as "most probable, given available evidence," never as certainty.

## README
Explains the why of the repo's existence and its place in the series, referencing
ADRs for detail instead of repeating it.

## Language
English for all prose. Test names in English per ADR-006.