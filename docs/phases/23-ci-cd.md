# Phase 23 — CI/CD

> Real Go continuous integration gates that run alongside the existing
> semantic-release JS pipeline.

## Deliverables

### 23.1 Existing state
- `.github/workflows/` already held `pr-build.yml`, `lint.yml`,
  `labeler.yml`, and `release.yml` — but these only exercise the Bun /
  semantic-release toolchain, never the Go module wiring.

### 23.2 New `ci.yml`
Four jobs run on any PR to `main`/`staging` and on every push to `staging`:
- **Build, vet & test** — `go build`, `go vet`, a strict `gofmt` check, and
  the full unit suite.
- **End-to-end** — boots temporary `postgres:17` and `redis:7` GitHub
  Actions service containers, then runs `go test ./tests/e2e/...` against
  `TEST_DATABASE_URL` (postgres at `localhost:5432`).
- **Vulnerability scan** — `govulncheck` action over the module.
- **Docker build** — Buildx with GHA layer cache to keep image builds fast.

The existing semantic-release `release.yml` still publishes the versioned
image to the registry on merge, so the pipeline is end-to-end.

## Verify
- Push the branch and confirm all four CI jobs go green on the PR.