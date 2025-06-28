#!/bin/bash
set -e

# Security Check Script
# Combines static analysis (golangci-lint) with vulnerability scanning (govulncheck)
# Exit codes: 0=success, 1=linting issues, 2=vulnerabilities found, 3=both

LINT_EXIT=0
VULN_EXIT=0

echo "🔍 Running static security analysis..."
if ! $(go env GOPATH)/bin/golangci-lint run --enable-only=gosec,depguard,forbidigo,contextcheck,containedctx,noctx,errcheck,bodyclose,rowserrcheck,sqlclosecheck; then
    echo "❌ Static security analysis found issues"
    LINT_EXIT=1
else
    echo "✅ Static security analysis passed"
fi

echo ""
echo "🔍 Running vulnerability scan..."
if ! $(go env GOPATH)/bin/govulncheck ./...; then
    echo "❌ Vulnerability scan found issues"
    VULN_EXIT=2
else
    echo "✅ Vulnerability scan passed"
fi

# Combine exit codes
EXIT_CODE=$((LINT_EXIT + VULN_EXIT))

case $EXIT_CODE in
    0) echo "✅ All security checks passed" ;;
    1) echo "❌ Static analysis issues found" ;;
    2) echo "❌ Vulnerabilities found" ;;
    3) echo "❌ Both static analysis issues and vulnerabilities found" ;;
esac

exit $EXIT_CODE