# 🔄 GitHub Actions/CI-CD Setup - Complete

## ✅ What's Been Added

### Workflows
1. **CI/CD Pipeline** (`.github/workflows/ci.yml`)
   - ✅ Run on every push to `main` and `develop`
   - ✅ Run on every pull request
   - Tests, builds, security scanning, coverage reports

2. **Database Migration** (`.github/workflows/migration.yml`)
   - Manual workflow for production/staging migrations
   - Supports environment selection (staging/production)
   - Slack notifications (optional)

3. **Release** (`.github/workflows/release.yml`)
   - Triggered by git tags (`v*`)
   - Builds binaries for Linux, macOS, Windows
   - Creates GitHub releases
   - Optional Docker Hub push

### Automation
4. **Dependabot Configuration** (`.github/dependabot.yml`)
   - Auto-updates Go dependencies weekly
   - Auto-updates Docker base images
   - Auto-updates GitHub Actions

5. **Code Owners** (`.github/CODEOWNERS`)
   - Define review requirements for critical files
   - Auto-assign reviewers

6. **CI/CD Documentation** (`.github/CI_CD_GUIDE.md`)
   - Complete setup instructions
   - Troubleshooting guide
   - Best practices

### Docker
7. **Dockerfile** - Multi-stage optimized build
8. **.dockerignore** - Optimize build context

---

## 🚀 Quick Start

### 1. Push to GitHub
```bash
git add .
git commit -m "ci: add github actions workflows"
git push origin main
```

Monitor at: `https://github.com/sezarsaman/wallet-balance-split/actions`

### 2. Configure Secrets (Optional)
Go to **Settings → Secrets and variables → Actions**

For Docker Hub push on release:
```
DOCKER_USERNAME = your_docker_username
DOCKER_PASSWORD = your_docker_token
```

### 3. Create a Release
```bash
git tag v1.0.0
git push origin v1.0.0
```

Monitor at: `https://github.com/sezarsaman/wallet-balance-split/releases`

---

## 📋 CI Pipeline Flow

```
Push/PR to main
    ↓
Tests (with postgres service)
    ↓
Build
    ↓
Code quality checks (fmt, vet, gosec)
    ↓
Coverage reports
    ↓
✅ Pass / ❌ Fail
```

---

## 🎯 Status Badges for README

Add to your README.md:

```markdown
## Status

[![CI/CD](https://github.com/sezarsaman/wallet-balance-split/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/sezarsaman/wallet-balance-split/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/sezarsaman/wallet-balance-split/branch/main/graph/badge.svg)](https://codecov.io/gh/sezarsaman/wallet-balance-split)
```

---

## 📁 Files Created/Modified

```
.github/
├── workflows/
│   ├── ci.yml              # Main CI pipeline
│   ├── migration.yml       # Database migration
│   └── release.yml         # Release builds
├── dependabot.yml          # Dependency updates
├── CODEOWNERS              # Code ownership
└── CI_CD_GUIDE.md          # Complete documentation

Dockerfile                  # Multi-stage build
.dockerignore               # Build optimization
```

---

## 🔐 Next Steps (Optional)

1. **Enable Branch Protection**
   - Require CI to pass before merging
   - Settings → Branches → Add rule

2. **Connect Codecov**
   - Sign up at codecov.io
   - Repository will auto-report coverage

3. **Setup Slack Notifications**
   - Add `SLACK_WEBHOOK` secret to enable notifications
   - See `.github/CI_CD_GUIDE.md` for details

4. **Enable Docker Hub Pushes**
   - Add Docker Hub credentials to secrets
   - Release workflow will auto-push images

---

## 🎓 Learn More

- [GitHub Actions Docs](https://docs.github.com/en/actions)
- [Dependabot Docs](https://docs.github.com/en/code-security/dependabot)
- [Docker Multi-Stage Builds](https://docs.docker.com/build/building/multi-stage/)
