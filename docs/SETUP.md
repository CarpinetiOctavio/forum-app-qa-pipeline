# Setup

## Prerequisites

```bash
go version     # 1.24 or higher (see backend/go.mod)
node --version # 18 or higher (matches .github/workflows/ci.yml)
npm --version
```

> The Go module was renamed from `tp06-testing` to `forum-app-ci-testing`
> when this repo was rebuilt as a portfolio piece. Any reference to the old
> name in git history is intentional and traceable to ADR-004.

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
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs
# Windows: https://nodejs.org/
```

## Clone and install

```bash
git clone https://github.com/CarpinetiOctavio/forum-app-ci-testing.git
cd forum-app-ci-testing

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

## What this repo does NOT need

Unlike the later repos in this series, this repo requires no external
infrastructure:
- No GitHub Secrets — the CI pipeline uses no external service (see ADR-004).
  Coverage is uploaded as an internal pipeline artifact, not to Codecov or
  any third-party reporting tool.
- No container registry, no cloud hosting account, no Render/AWS/GCP
  configuration — those belong to `forum-app-cloud-deploy`.
- No SonarCloud token — static analysis is out of this repo's scope (see
  `docs/rules/testing.md`); it belongs to `forum-app-qa-pipeline`.

Cloning the repo and running the two install commands above is the entire
setup.
