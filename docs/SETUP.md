# Setup

## Prerequisites

```bash
go version     # 1.24 or higher (see backend/go.mod)
node --version # 24 (matches .github/workflows/ci.yml)
npm --version
```

> The Go module is `forum-app-qa-pipeline`. This repository started as a copy
> of `forum-app-ci-testing@v1.1.0`, so earlier names appear in its history
> (see ADR-000).

**Installing Go:**
```bash
# macOS
brew install go
# Ubuntu/Debian
sudo apt install golang-go
# Windows: https://go.dev/dl/
```

**Installing Node.js:**
```bash
# macOS
brew install node
# Ubuntu/Debian
curl -fsSL https://deb.nodesource.com/setup_24.x | sudo -E bash -
sudo apt-get install -y nodejs
# Windows: https://nodejs.org/
```

## Clone and install

```bash
git clone https://github.com/CarpinetiOctavio/forum-app-qa-pipeline.git
cd forum-app-qa-pipeline

cd backend && go mod download && cd ..
cd frontend && npm install && cd ..
```

## Running locally

```bash
# Terminal 1 — backend
cd backend && go run cmd/api/main.go
# http://localhost:8080

# Terminal 2 — frontend
cd frontend && npm start
# http://localhost:3000
```

## Database schema

`backend/internal/database/database.go` creates `backend/database.db` on first
run (`InitDB("./database.db")` in `main.go`), using `CREATE TABLE IF NOT
EXISTS` so re-running the app never wipes existing data. Three tables:

```
users                          posts                           comments
├── id (PK)                    ├── id (PK)                     ├── id (PK)
├── email (UNIQUE)              ├── title                       ├── post_id (FK → posts)
├── password                    ├── content                     ├── user_id (FK → users)
├── username                    ├── user_id (FK → users)         ├── content
└── created_at                  └── created_at                  └── created_at
```

Both foreign keys that reference `users`, and the one referencing `posts`,
are declared `ON DELETE CASCADE`. Deleting a user deletes their posts *and*
every comment on those posts; deleting a post deletes its comments. This
matters for local testing: seeding a user, a few posts, and comments, then
deleting the user, is enough to verify the cascade — no manual cleanup of
orphaned rows is needed, and none is possible to forget. Indexes exist on
all three foreign-key columns (`idx_posts_user_id`,
`idx_comments_post_id`, `idx_comments_user_id`).

Tests never touch this file — every repository is mocked in the service-layer
tests, so `database.db` is only ever created by actually running the app
(`go run cmd/api/main.go`) or the E2E suite.

## Frontend service layer

Components never call `axios` directly. `frontend/src/services/authService.ts`
and `postService.ts` wrap every HTTP call and return typed data
(`response.data`, not the full Axios response). A component only ever calls
`authService.login(credentials)` or `postService.createPost(data, userId)` —
it has no idea `axios` is involved, what the exact URL is, or that HTTP is
happening at all.

This pays off in two concrete ways: changing the API base URL or switching
HTTP libraries touches two files instead of every component that makes a
request, and testing a component only requires mocking `axios` once (via
`jest.mock('axios')`) rather than stubbing HTTP calls scattered across
component code.

## What the CI pipeline needs

- **A SonarCloud project and token.** The project key is
  `CarpinetiOctavio_forum-app-qa-pipeline2` in the `carpinetioctavio`
  organization (`sonar-project.properties`; ADR-002 explains the `2`). The
  token is the repository secret `SONAR_TOKEN`. The organization is on the
  Free plan, which shows a quality gate only for `main` and for pull requests
  into `main` (ADR-002).
- **No container registry, cloud hosting account or deploy credentials** —
  those belong to `forum-app-cloud-deploy`.

Running the app and the tests locally needs only the clone and the two install
commands above.
