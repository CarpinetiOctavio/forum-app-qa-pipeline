# COMMANDS.md — forum-app-qa-pipeline

Commands that work against this repo's current state. SonarCloud runs only in
CI (it needs the `SONAR_TOKEN` secret; see `SETUP.md`).

## Backend

```bash
cd backend

# Run the server (http://localhost:8080)
go run cmd/api/main.go

# Run all unit tests
go test ./tests/services/... -v

# Run with coverage (measures internal/services/ only — see ADR-001)
go test ./tests/services/... -v -cover -coverpkg=./internal/services/...

# Generate an HTML coverage report
go test ./tests/services/... -coverprofile=coverage.out -coverpkg=./internal/services/...
go tool cover -html=coverage.out

# Coverage summary in terminal
go tool cover -func=coverage.out
```

## Frontend

```bash
cd frontend

# Run the dev server (http://localhost:3000)
npm start

# Run tests once
npm test -- --watchAll=false

# Run tests with coverage (fails below the 70% threshold in package.json — see ADR-001)
npm test -- --coverage --watchAll=false
```

## Cypress (E2E)

With the backend (`go run cmd/api/main.go`) and the frontend (`npm start`)
running in two other terminals:

```bash
cd frontend

# Run every spec headless, as CI does
npx cypress run

# Only the spec that uses the real backend and SQLite (see ADR-003)
npx cypress run --spec cypress/e2e/delete-comment-real.cy.ts
```

## Git / branching

```bash
# Feature branches merge into staging, staging merges into main
git checkout -b feature/<name> staging

# Trigger the CI pipeline manually (empty commit)
git commit --allow-empty -m "trigger pipeline"
git push
```