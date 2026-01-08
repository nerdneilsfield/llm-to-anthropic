projectname := "llm-to-anthropic"

# List all available commands
default:
    @just --list

# Build Golang binary
build:
    go build -ldflags "-X main.version=$(git describe --abbrev=0 --tags)" -o {{projectname}}

# Install Golang binary
install:
    go install -ldflags "-X main.version=$(git describe --abbrev=0 --tags)"

# Run application
run:
    go run -ldflags "-X main.version=$(git describe --abbrev=0 --tags)" main.go

# Install build dependencies
bootstrap:
    go generate -tags tools tools/tools.go

# Run tests with coverage
test: clean
    go test --cover -parallel=1 -v -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out | sort -rnk3

# Clean build artifacts
clean:
    rm -rf coverage.out dist {{projectname}} {{projectname}}.exe

# Show test coverage
cover:
    go test -v -race $(go list ./... | grep -v /vendor/) -v -coverprofile=coverage.out
    go tool cover -func=coverage.out

# Format Go files
fmt:
    gofumpt -w .
    gci write .

# Run linter
lint:
    golangci-lint run -c .golang-ci.yml

# Test release
release-test:
    goreleaser release  --snapshot --clean

# Run pre-commit hooks (commented out)
# pre-commit:
#     pre-commit run --all-files

# Build for all platforms
build-all:
    goreleaser build --snapshot --clean
