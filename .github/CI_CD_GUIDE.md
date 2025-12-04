# 🚀 CI/CD Pipeline Documentation

## Workflows Overview

### 1. **CI/CD Pipeline** (`.github/workflows/ci.yml`)
Triggered on every push to `main`/`develop` and pull requests.

**Steps:**
- ✅ Code checkout
- ✅ Go setup & dependency download
- ✅ Run migrations (if migration script exists)
- ✅ Run tests with coverage (`go test -v -race -coverprofile=coverage.out`)
- ✅ Upload coverage to Codecov
- ✅ Build application binary
- ✅ Code formatting check (`gofmt`)
- ✅ Static analysis (`go vet`)
- ✅ Security scanning (`gosec`)

**Build Requirements:**
- PostgreSQL 15-alpine (auto-started as service)
- Go 1.21+

**Outputs:**
- Coverage reports (Codecov)
- Security scan results (GitHub Code Scanning)

---

### 2. **Database Migration Workflow** (`.github/workflows/migration.yml`)
Manual workflow for production/staging database migrations.

**Trigger:** Manual via GitHub Actions UI
**Inputs:**
- `environment`: staging or production

**Steps:**
- 📝 Run migrations using stored secrets
- ✅ Verify migrations completed
- 📢 Send Slack notification (optional)

**Required Secrets:**
```
DB_HOST_staging        # Database host for staging
DB_PORT_staging        # Database port for staging
DB_USER_staging        # Database user for staging
DB_PASSWORD_staging    # Database password for staging
DB_NAME_staging        # Database name for staging
DB_HOST_production     # Database host for production
DB_PORT_production     # Database port for production
DB_USER_production     # Database user for production
DB_PASSWORD_production # Database password for production
DB_NAME_production     # Database name for production
SLACK_WEBHOOK         # (Optional) Slack webhook URL
```

---

### 3. **Release Workflow** (`.github/workflows/release.yml`)
Triggered when a git tag matching `v*` is pushed.

**Trigger:** `git push origin v1.0.0`

**Steps:**
- 📦 Build binaries for multiple platforms:
  - Linux (amd64)
  - macOS (amd64, arm64)
  - Windows (amd64)
- 🔐 Generate SHA256 checksums
- 🏷️ Create GitHub Release with binaries
- 🐳 Build and push Docker image (if Docker Hub secrets configured)

**Required Secrets (for Docker Hub push):**
```
DOCKER_USERNAME  # Docker Hub username
DOCKER_PASSWORD  # Docker Hub password or token
```

---

## Setup Instructions

### 1. Configure GitHub Secrets

Go to **Settings → Secrets and variables → Actions** and add:

#### For CI/CD:
No secrets required (uses default GitHub token)

#### For Database Migration:
```bash
# Navigate to repo settings
# Add the following secrets based on your environments
```

#### For Release (Optional):
```bash
# Docker Hub credentials
DOCKER_USERNAME=your_username
DOCKER_PASSWORD=your_access_token
```

### 2. Environment Variables

The workflows use environment variables from `.env` file locally.
For GitHub Actions, these should be defined in:
- Workflow files (hardcoded non-sensitive values)
- GitHub Secrets (sensitive values like passwords)

### 3. Test the CI Pipeline

```bash
# Make a change and push to develop
git add .
git commit -m "test: CI pipeline"
git push origin develop

# Monitor at: https://github.com/sezarsaman/wallet-balance-split/actions
```

### 4. Create a Release

```bash
# Create a tag
git tag v1.0.0
git push origin v1.0.0

# Monitor at: https://github.com/sezarsaman/wallet-balance-split/releases
```

---

## Local Testing

### Test migrations locally:
```bash
go run ./cmd/migrate.go
```

### Run tests locally:
```bash
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # View coverage report
```

### Build locally:
```bash
go build -o ./bin/wallet-api ./cmd/main.go
```

### Check code formatting:
```bash
gofmt -l .
gofmt -w .  # Auto-fix
```

### Run security scan:
```bash
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec ./...
```

---

## Troubleshooting

### ❌ "Database connection failed" in CI
- Ensure PostgreSQL service in CI is healthy
- Check DB credentials match service definition
- Verify migrations are idempotent

### ❌ "Test failures in CI but pass locally"
- Check for race conditions: `go test -race`
- Verify environment variables are set in workflow
- Check database state is clean (migrations should be idempotent)

### ❌ "Security scan keeps failing"
- Review gosec findings: `gosec ./...`
- Suppress false positives with `#nosec` comments where appropriate
- Update vulnerable dependencies: `go get -u ./...`

### ❌ "Docker build fails in release"
- Verify Dockerfile paths are correct
- Check `.dockerignore` is not excluding required files
- Ensure binary is correctly built in multi-stage

---

## Status Badges

Add these to your README.md:

```markdown
[![CI/CD](https://github.com/sezarsaman/wallet-balance-split/actions/workflows/ci.yml/badge.svg)](https://github.com/sezarsaman/wallet-balance-split/actions/workflows/ci.yml)
[![Release](https://github.com/sezarsaman/wallet-balance-split/actions/workflows/release.yml/badge.svg)](https://github.com/sezarsaman/wallet-balance-split/actions/workflows/release.yml)
[![codecov](https://codecov.io/gh/sezarsaman/wallet-balance-split/branch/main/graph/badge.svg)](https://codecov.io/gh/sezarsaman/wallet-balance-split)
```

---

## Best Practices

1. **Always run tests before pushing**: `go test -v ./...`
2. **Use semantic versioning**: `v1.0.0`, `v1.1.0`, etc.
3. **Keep migrations idempotent**: Use `CREATE TABLE IF NOT EXISTS`
4. **Tag releases from main branch**: Ensure tests pass first
5. **Monitor action runs**: Check GitHub Actions tab regularly
6. **Document breaking changes**: Include in release notes

---

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Testing Documentation](https://golang.org/doc/effective_go#testing)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
