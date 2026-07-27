# Command Reference

## Backend (Go)

```bash
cd backend

# Run the server
go run cmd/api/main.go

# Build a binary
go build -o app cmd/api/main.go

# Run all unit tests, verbose
go test ./tests/services/... -v

# Run with coverage
go test ./tests/services/... -v -cover

# Coverage profile + HTML report
go test ./tests/services/... -coverprofile=coverage.out
go tool cover -html=coverage.out
go tool cover -func=coverage.out

# Run a single test
go test ./tests/services/... -v -run TestRegister_EmptyEmail

# Tidy dependencies
go mod tidy
```

## Frontend (React)

```bash
cd frontend

# Start the dev server
npm start

# Run all tests once (CI mode)
npm test -- --watchAll=false

# Run with coverage
npm test -- --coverage --watchAll=false

# Run a single test file
npm test -- Login.test.tsx --watchAll=false

# Production build
npm run build
```

## Both suites

```bash
cd backend && go test ./tests/services/... -v && cd ../frontend && npm test -- --watchAll=false
```

## CI

The pipeline (`.github/workflows/ci.yml`) runs automatically on every push or
pull request to `main`, `master`, or `develop`. To trigger a run without a
code change:

```bash
git commit --allow-empty -m "chore: trigger pipeline"
git push origin main
```

View results at:
`https://github.com/CarpinetiOctavio/forum-app-ci-testing/actions`

## Git history evidence

Used to capture the branching-model screenshots
(`docs/screenshots/09-git-history-before-branching.png` and
`10a`/`10b-git-history-after-branching-*.png`):

```bash
git log --graph --all --decorate --format="%C(yellow)%h%C(reset) %C(cyan)%ad%C(reset)%C(auto)%d%C(reset) %s %C(green)(%an)%C(reset)" --date=short
```

`--all` shows every local branch, not just the checked-out one; the `%d` in
`--format` is what actually prints ref/tag labels (`HEAD -> ...`, `origin/main`,
`v1.0.0`) inline on the graph — `--decorate` alone has no effect once a custom
`--format` is supplied.

## Verifying test isolation

These are manual checks, not automated ones — they demonstrate that the mocking
strategy (see ADR-003) actually isolates tests from the dependencies it claims
to isolate them from, rather than just asserting that it does.

**Backend tests do not depend on a real database:**
```bash
rm -f backend/database.db
cd backend && go test ./tests/services/... -v
# Passes identically with no database.db present — the Repository layer is
# mocked (tests/mocks/), the real SQLite file is never touched
```

**Frontend tests do not depend on a running backend:**
```bash
# Do not start the backend for this — go straight to the frontend
cd frontend && npm test -- --watchAll=false
# Passes identically with no backend process running on :8080 — axios is
# mocked (src/__mocks__/axios.ts), no HTTP request ever leaves the process
```

**Tests are deterministic across repeated runs:**
```bash
cd backend
for i in {1..10}; do go test ./tests/services/... -v; done
# All 10 runs produce the same 24/24 pass result — mocks return fixed,
# pre-configured values, so there is no shared or accumulating state between runs
```

## Troubleshooting

**Backend won't start (port in use):**
```bash
lsof -i :8080
kill -9 <PID>
```

**Frontend won't start / stale dependencies:**
```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
```

**`npm ci` fails in CI with a lockfile mismatch:**
See ADR-005 — regenerate the lockfile locally and commit it:
```bash
cd frontend
rm package-lock.json
npm install
git add package-lock.json
git commit -m "fix: regenerate package-lock.json"
```

**Backend tests fail after a fresh clone:**
```bash
# Confirm tests don't depend on a local database file
rm -f backend/database.db
cd backend && go test ./tests/services/... -v
# Should pass identically — tests mock the repository layer (see ADR-003)
```
