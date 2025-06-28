#!/bin/bash
set -e

# Custom Security Linting Script
# Detects security issues that standard linters might miss

echo "🔍 Running comprehensive security analysis..."

EXIT_CODE=0

# 1. Credential Detection
echo "🔐 Checking for credential logging..."
if grep -r "username\|password\|secret\|token\|key" --include="*.go" . | grep -E "log\.|Log\.|Info\(" | grep -v "test"; then
    echo "❌ Found potential credential logging"
    EXIT_CODE=1
else
    echo "✅ No credential logging detected"
fi

# 2. Unsafe HTTP Patterns
echo "🌐 Checking for unsafe HTTP configurations..."
if grep -r "ListenAndServe.*nil" --include="*.go" .; then
    echo "❌ Found unsafe HTTP server configuration"
    EXIT_CODE=1
else
    echo "✅ HTTP configurations are secure"
fi

# 3. Path Traversal Risks
echo "📁 Checking for path traversal vulnerabilities..."
if grep -r "\.\." --include="*.go" . | grep -E "os\.|ioutil\.|filepath\." | grep -v "test" | grep -v "Clean"; then
    echo "❌ Found potential path traversal risk"
    EXIT_CODE=1
else
    echo "✅ No path traversal risks detected"
fi

# 4. Error Information Disclosure
echo "🚫 Checking for information disclosure in errors..."
if grep -r "fmt\.Errorf.*%s.*body\|fmt\.Errorf.*%v.*response" --include="*.go" . | grep -v "test"; then
    echo "❌ Found potential information disclosure in error messages"
    EXIT_CODE=1
else
    echo "✅ Error messages are sanitized"
fi

# 5. Race Condition Patterns
echo "🏁 Checking for potential race conditions..."
if grep -r "var.*=.*New\|globalLogger\s*=" --include="*.go" . | grep -v "mutex\|sync\|atomic" | grep -v "test" | grep -v "once"; then
    echo "⚠️  Found potential race condition patterns (manual review needed)"
    # Don't fail on this one as it needs manual review
fi

# 6. Nil Pointer Dereference Risks
echo "💥 Checking for nil pointer dereference risks..."
if grep -r "\*[a-zA-Z_][a-zA-Z0-9_]*\." --include="*.go" . | grep -v "test" | head -5; then
    echo "⚠️  Found pointer dereferences (check for nil validation)"
    # Don't fail on this one as it needs context analysis
fi

# 7. Unsafe File Permissions
echo "🔒 Checking for unsafe file permissions..."
if grep -r "0777\|0666\|0755" --include="*.go" . | grep -v "test" | grep -v "0755"; then
    echo "❌ Found potentially unsafe file permissions"
    EXIT_CODE=1
else
    echo "✅ File permissions are secure"
fi

# 8. SQL Injection Patterns (backup check)
echo "💉 Checking for SQL injection patterns..."
if grep -r "fmt\.Sprintf.*SELECT\|fmt\.Sprintf.*INSERT\|fmt\.Sprintf.*UPDATE\|fmt\.Sprintf.*DELETE" --include="*.go" .; then
    echo "❌ Found potential SQL injection vulnerability"
    EXIT_CODE=1
else
    echo "✅ No SQL injection patterns detected"
fi

# 9. Missing Input Validation
echo "🔍 Checking for missing input validation..."
if grep -r "r\.Form\|r\.PostForm\|r\.URL\.Query" --include="*.go" . | grep -v "validate\|check\|sanitize" | grep -v "test"; then
    echo "⚠️  Found form inputs without apparent validation (manual review needed)"
fi

# 10. Hardcoded Secrets
echo "🔑 Checking for hardcoded secrets..."
if grep -rE "(password|secret|key|token)\s*[:=]\s*['\"][^'\"]{8,}" --include="*.go" . | grep -v "test" | grep -v "example"; then
    echo "❌ Found potential hardcoded secrets"
    EXIT_CODE=1
else
    echo "✅ No hardcoded secrets detected"
fi

# 11. Goroutine Leak Detection
echo "🔄 Checking for potential goroutine leaks..."
if grep -r "go\s\+[^}]*(" --include="*.go" . | grep -v "context\|cancel\|timeout\|done\|defer.*Stop\|stopChan" | grep -v "test"; then
    echo "⚠️  Found goroutines without apparent lifecycle management (manual review needed)"
    echo "    Consider using context.WithCancel(), timeout patterns, or proper Stop() methods"
fi

# 12. CSRF Protection Missing
echo "🛡️ Checking for missing CSRF protection..."
if grep -r "method.*post\|Method.*POST" --include="*.go" . | grep -v "csrf\|token" | grep -v "test"; then
    echo "⚠️  Found POST handlers without apparent CSRF protection (manual review needed)"
    echo "    Consider adding CSRF middleware for state-changing operations"
fi

# 13. Missing Authentication Checks
echo "🔐 Checking for missing authentication..."
if grep -r "http\.Handler\|HandlerFunc\|HandleFunc" --include="*.go" . | grep -v "auth\|login\|authenticate" | grep -v "test" | head -3; then
    echo "⚠️  Found HTTP handlers without apparent authentication (manual review needed)"
fi

# 14. Resource Leak Detection - File Operations
echo "📁 Checking for unclosed file operations..."
FILE_OPENS=$(grep -rn "os\.Open\|os\.Create\|os\.OpenFile\|ioutil\.ReadFile\|ioutil\.WriteFile" --include="*.go" . | grep -v "test" | wc -l)
FILE_CLOSES=$(grep -rn "\.Close()\|defer.*Close" --include="*.go" . | grep -v "test" | wc -l)
if [ $FILE_OPENS -gt 0 ] && [ $FILE_CLOSES -lt $FILE_OPENS ]; then
    echo "⚠️  Found $FILE_OPENS file operations but only $FILE_CLOSES close calls (check for resource leaks)"
    echo "    Each file open should have a corresponding defer file.Close()"
else
    echo "✅ File operations appear to have proper cleanup"
fi

# 15. Resource Leak Detection - HTTP Response Bodies
echo "🌐 Checking for unclosed HTTP response bodies..."
HTTP_RESPONSES=$(grep -rn "http\.Client\|httpClient\.Do\|http\.Get\|http\.Post" --include="*.go" . | grep -v "test" | wc -l)
BODY_CLOSES=$(grep -rn "defer.*Body\.Close\|resp\.Body\.Close" --include="*.go" . | grep -v "test" | wc -l)
if [ $HTTP_RESPONSES -gt 0 ] && [ $BODY_CLOSES -lt $HTTP_RESPONSES ]; then
    echo "⚠️  Found $HTTP_RESPONSES HTTP operations but only $BODY_CLOSES body close calls"
    echo "    Each HTTP response should have: defer resp.Body.Close()"
else
    echo "✅ HTTP response bodies appear to have proper cleanup"
fi

# 16. Resource Leak Detection - Database Resources
echo "🗄️ Checking for unclosed database resources..."
DB_QUERIES=$(grep -rn "\.Query\|\.QueryRow\|\.Prepare" --include="*.go" . | grep -v "test" | wc -l)
DB_CLOSES=$(grep -rn "defer.*Close\|rows\.Close\|stmt\.Close" --include="*.go" . | grep -v "test" | wc -l)
if [ $DB_QUERIES -gt 0 ] && [ $DB_CLOSES -lt $DB_QUERIES ]; then
    echo "⚠️  Found $DB_QUERIES database operations but only $DB_CLOSES close calls"
    echo "    Database rows and statements should have: defer rows.Close() or defer stmt.Close()"
else
    echo "✅ Database resources appear to have proper cleanup"
fi

# 17. Missing Context Cancellation
echo "⏰ Checking for missing context cancellation..."
CONTEXT_CREATES=$(grep -rn "context\.WithCancel\|context\.WithTimeout\|context\.WithDeadline" --include="*.go" . | grep -v "test" | wc -l)
CONTEXT_CANCELS=$(grep -rn "defer.*cancel\|cancel()" --include="*.go" . | grep -v "test" | wc -l)
if [ $CONTEXT_CREATES -gt 0 ] && [ $CONTEXT_CANCELS -lt $CONTEXT_CREATES ]; then
    echo "⚠️  Found $CONTEXT_CREATES context creations but only $CONTEXT_CANCELS cancel calls"
    echo "    Each context creation should have: defer cancel()"
else
    echo "✅ Context cancellation appears to be properly handled"
fi

# 18. Ticker/Timer Resource Leaks
echo "⏲️ Checking for unclosed tickers and timers..."
TICKER_CREATES=$(grep -rn "time\.NewTicker\|time\.NewTimer" --include="*.go" . | grep -v "test" | wc -l)
TICKER_STOPS=$(grep -rn "\.Stop()\|defer.*Stop" --include="*.go" . | grep -v "test" | wc -l)
if [ $TICKER_CREATES -gt 0 ] && [ $TICKER_STOPS -lt $TICKER_CREATES ]; then
    echo "⚠️  Found $TICKER_CREATES ticker/timer creations but only $TICKER_STOPS stop calls"
    echo "    Each ticker/timer should have: defer ticker.Stop()"
else
    echo "✅ Tickers and timers appear to have proper cleanup"
fi

echo ""
if [ $EXIT_CODE -eq 0 ]; then
    echo "✅ Security analysis completed successfully"
else
    echo "❌ Security issues found - review required"
fi

exit $EXIT_CODE