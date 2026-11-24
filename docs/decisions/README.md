# Decisions

Each Architecture Decision Record (ADR) states one decision: its context, the alternatives considered and the reasons for the choice. This index only says what each record decides; the reasoning lives in the record itself. Later corrections are added as dated amendments at the end of the record; the original text is not rewritten. Decisions this repository builds on from [`forum-app-ci-testing`](https://github.com/CarpinetiOctavio/forum-app-ci-testing/tree/main/docs/decisions) are linked from the record that uses them, not copied here.

| ADR | Area | What it decides |
|---|---|---|
| [000](ADR-000-starting-from-ci-testing-v1.1.0.md) | Method | The repository starts as a copy of `forum-app-ci-testing@v1.1.0`, not as a continuation of the original TP7 submission (`forum-app-qa-pipeline-legacy`). Amended 2026-11-12. |
| [001](ADR-001-coverage-strategy.md) | Coverage | A 70% coverage gate in backend and frontend, measured only over the backend Services layer and the frontend components and services. |
| [002](ADR-002-sonarcloud.md) | Static analysis | SonarCloud through CI-based analysis, and the `sonar.sources`/`sonar.tests` configuration that makes it analyze the whole repository. Amended 2026-11-12 (the job waits for the quality gate only on the path to `main`, within the Free plan's limits). |
| [003](ADR-003-cypress.md) | E2E testing | Cypress against mocked responses, except `DeleteComment`, whose two tests run against the real backend and SQLite. Amended 2026-11-12 (the frontend is started in its own step and waited for with `wait-on`). Includes a diagram. |
| [004](ADR-004-pipeline-extension.md) | Pipeline | SonarCloud and Cypress run after the unit-test jobs, in parallel with each other; the coverage gate lives inside the test jobs; pushes to `staging` and `main` run the pipeline. Amended 2026-11-12. Includes a diagram. |
| [005](ADR-005-edit-functionality.md) | Application | Posts and comments can be edited, with the authorship check in the service layer, not in SQL. Includes a diagram. |
| [006](ADR-006-sqlite-and-unenforced-foreign-keys.md) | Database | SQLite as the database engine, recorded retroactively with the reasons given in TP6 and the alternatives set aside then. Its foreign keys are declared but not enforced: a defect that comes from `forum-app-ci-testing`, recorded and not fixed here. |
