#!/bin/bash

# 🚀 CI/CD Setup Verification Checklist

echo "╔════════════════════════════════════════════════════════════╗"
echo "║    🚀 GitHub Actions CI/CD Verification Checklist         ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counters
TOTAL=0
PASSED=0

# Function to check file
check_file() {
    local file=$1
    local desc=$2
    TOTAL=$((TOTAL + 1))
    
    if [ -f "$file" ]; then
        echo -e "${GREEN}✓${NC} $desc ($file)"
        PASSED=$((PASSED + 1))
        return 0
    else
        echo -e "${RED}✗${NC} $desc ($file)"
        return 1
    fi
}

# Function to check directory
check_dir() {
    local dir=$1
    local desc=$2
    TOTAL=$((TOTAL + 1))
    
    if [ -d "$dir" ]; then
        echo -e "${GREEN}✓${NC} $desc ($dir)"
        PASSED=$((PASSED + 1))
        return 0
    else
        echo -e "${RED}✗${NC} $desc ($dir)"
        return 1
    fi
}

echo "📁 File Structure:"
echo "─────────────────────────────────────────────────────────────"

check_dir ".github" "GitHub workflows directory"
check_dir ".github/workflows" "Workflows subdirectory"

echo ""
echo "📝 Workflow Files:"
echo "─────────────────────────────────────────────────────────────"

check_file ".github/workflows/ci.yml" "CI/CD pipeline"
check_file ".github/workflows/migration.yml" "Database migration"
check_file ".github/workflows/release.yml" "Release workflow"

echo ""
echo "⚙️  Configuration Files:"
echo "─────────────────────────────────────────────────────────────"

check_file ".github/dependabot.yml" "Dependabot configuration"
check_file ".github/CODEOWNERS" "Code owners"

echo ""
echo "📚 Documentation:"
echo "─────────────────────────────────────────────────────────────"

check_file ".github/CI_CD_GUIDE.md" "CI/CD complete guide"
check_file ".github/SETUP.md" "Quick start guide"
check_file ".github/IMPLEMENTATION_SUMMARY.md" "Implementation summary"

echo ""
echo "🐳 Docker Files:"
echo "─────────────────────────────────────────────────────────────"

check_file "Dockerfile" "Multi-stage Dockerfile"
check_file ".dockerignore" "Docker ignore file"

echo ""
echo "🔍 Validation Checks:"
echo "─────────────────────────────────────────────────────────────"

TOTAL=$((TOTAL + 1))
if go version >/dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Go installed"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}✗${NC} Go not installed"
fi

TOTAL=$((TOTAL + 1))
if docker --version >/dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Docker installed"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}✗${NC} Docker not installed"
fi

TOTAL=$((TOTAL + 1))
if git --version >/dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Git installed"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}✗${NC} Git not installed"
fi

echo ""
echo "═════════════════════════════════════════════════════════════"
echo "Summary: ${PASSED}/${TOTAL} checks passed"
echo "═════════════════════════════════════════════════════════════"

if [ $PASSED -eq $TOTAL ]; then
    echo -e "${GREEN}✓ All systems ready for GitHub!${NC}"
    echo ""
    echo "Next steps:"
    echo "1. git add ."
    echo "2. git commit -m 'ci: add github actions workflows'"
    echo "3. git push origin main"
    echo ""
    echo "Then visit: https://github.com/sezarsaman/wallet-balance-split/actions"
    exit 0
else
    echo -e "${YELLOW}⚠ Some checks failed. Please review above.${NC}"
    exit 1
fi
