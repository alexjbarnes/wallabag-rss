# Wallabag RSS Tool Justfile

# Default recipe
default: build

# Install required tools
install-tools:
    go install github.com/a-h/templ/cmd/templ@latest
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    go install mvdan.cc/gofumpt@latest
    go install golang.org/x/tools/cmd/goimports@latest
    go install golang.org/x/vuln/cmd/govulncheck@latest

# Generate templ files
generate:
    templ generate ./views

# Build the application
build: generate
    go build -o wallabag-rss-tool .

# Run the application
run: build
    ./wallabag-rss-tool

# Run in development mode with Air hot reload
dev: install-tools
    air

# Run tests
test: generate
    go test -v ./...

# Run tests with coverage
test-coverage: generate
    go test -cover ./...

# Generate detailed coverage report (excluding mocks and generated templates)
coverage-report: generate
    go test -coverprofile=coverage.out -coverpkg=./... ./...
    grep -v "_templ.go" coverage.out | grep -v "mocks/" > coverage_filtered.out
    go tool cover -html=coverage_filtered.out -o coverage.html
    @echo "📊 Coverage report generated: coverage.html"
    @go tool cover -func=coverage_filtered.out | grep "total:"

# Show coverage percentage only (excluding mocks and generated files)  
coverage-check:
    @templ generate ./views > /dev/null 2>&1
    @echo "Running tests and generating coverage..."
    @timeout 120 go test -coverprofile=coverage.out -coverpkg=./... ./... > /dev/null 2>&1 || true
    @if [ -f coverage.out ]; then \
        go tool cover -func=coverage.out | grep -E "^wallabag-rss-tool/(pkg|main)" | grep -v "_templ.go" | grep -v "/mocks/" | awk 'BEGIN{sum=0;count=0} {if(NF>=3 && $3!="") {sum+=substr($3,1,length($3)-1); count++}} END {if(count>0) printf "%.1f%%\n", sum/count; else print "0.0%"}'; \
    else \
        echo "0.0%"; \
    fi

# Clean build artifacts
clean:
    rm -f wallabag-rss-tool
    rm -f views/*_templ.go

# Format code
fmt:
    go fmt ./...
    templ fmt ./views

# Comprehensive linting: code quality + security + vulnerabilities + custom patterns
lint:
    @echo "🔍 Running comprehensive code analysis..."
    @echo "📋 1/4: Static code analysis & quality checks..."
    golangci-lint run
    @echo "🔒 2/4: Security vulnerability scanning..."
    govulncheck ./...
    @echo "🎯 3/4: Custom security pattern detection..."
    ./scripts/security-lint.sh
    @echo "✅ 4/4: Analysis complete - All checks passed!"

# Fix linting issues automatically
lint-fix:
    gofumpt -l -w .
    goimports -l -w .
    golangci-lint run --fix

# Watch for changes and regenerate templ files
watch:
    templ generate --watch --path=./views