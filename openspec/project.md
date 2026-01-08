# Project Context

## Purpose
A Go template project that provides a standardized starting point for Go CLI applications. It includes best practices for project structure, build tooling, testing, and release management.

## Tech Stack
- **Language**: Go 1.23.2
- **CLI Framework**: Cobra (github.com/spf13/cobra)
- **Logging**: Zap (go.uber.org/zap) + shlogin logger (github.com/nerdneilsfield/shlogin)
- **Build & Release**: GoReleaser
- **Containerization**: Docker
- **Linting**: golangci-lint with gofmt, goimports, govet, revive, errcheck, staticcheck, gosimple, ineffassign
- **Formatting**: gofumpt, gci

## Project Conventions

### Code Style
- **Formatting**: Use `gofumpt -w .` and `gci write .` to format code
- **Linting**: Run `golangci-lint run -c .golangci.yml` before committing
- **Import Organization**: Sorted and grouped using gci
- **Error Handling**: Always return errors; use `t.Fatalf` or `t.Errorf` for test failures
- **Naming**: Standard Go conventions - PascalCase for exported, camelCase for unexported
- **Comments**: Document exported functions and methods with comments describing behavior and error conditions

### Architecture Patterns
- **Project Layout**: Standard Go layout
  - `cmd/` - CLI command definitions using Cobra
  - `internal/` - Internal application code (not importable by external packages)
  - `pkg/` - Library code that can be used by external packages
- **CLI Structure**: Root command with subcommands using Cobra framework
- **Logger**: Singleton pattern using `loggerPkg.GetLogger()`
- **Graceful Shutdown**: Signal handling for SIGINT and SIGTERM with cleanup
- **Version Info**: Injected via ldflags (version, buildTime, gitCommit)
- **Build Tags**: Use build tags for tool dependencies (`go generate -tags tools tools/tools.go`)

### Testing Strategy
- **Framework**: Standard Go `testing` package
- **Test Naming**: `Test<FunctionName>` pattern (e.g., `TestRowWiseMatrix_Add`)
- **Coverage**: Run `make test` or `just test` to execute tests with coverage reports
- **Parallel Tests**: Tests run with `-parallel=1` flag
- **Race Detection**: Use `make cover` or `just cover` for race-enabled testing
- **Golden Pattern**: Compare results against expected values using helper methods like `Equals()`
- **Error Testing**: Always check for errors in tests using `if err != nil`

### Git Workflow
- **Branching**: Main development on `master` branch
- **Commit Messages**: Follow conventional commits pattern (inferred from goreleaser config):
  - `feat: new features`
  - `fix: bug fixes`
  - `docs: documentation changes`
  - `deps: dependency updates`
  - `chore: maintenance tasks`
- **Versioning**: Git tags for releases (e.g., v1.0.0)
- **Release**: GoReleaser automates GitHub releases with changelog grouping

## Domain Context
This is a foundational template project, not a domain-specific application. AI assistants should:
- Understand that `internal/call/` contains example matrix operations (addition, subtraction, multiplication)
- Recognize that `pkg/example/` demonstrates matrix operations and is intended as sample code
- Note that the project structure is designed to be copied and adapted for new CLI applications

## Important Constraints
- **CGO**: Disabled (CGO_ENABLED=0) for static builds
- **Cross-Platform**: Must support Linux, Darwin, Windows, FreeBSD across 386, amd64, arm64 architectures
- **Static Builds**: Linux builds are fully static for maximum portability
- **Go Version**: Minimum Go 1.23.2 required
- **No Vendor**: Dependencies managed via go.mod (not vendored)

## External Dependencies
- **shlogin logger**: Internal logging package (github.com/nerdneilsfield/shlogin) providing structured logging with verbose mode
- **GitHub**: Used for releases, issues, and container registry (ghcr.io/nerdneilsfield/go-template)
- **Docker Hub**: Container images published to docker.com (nerdneils/go-template)
- **Dependabot**: Automated dependency updates via .github/dependabot.yml

## Build Commands
- `make build` / `just build` - Build binary for current platform
- `make install` / `just install` - Install binary to GOPATH/bin
- `make test` / `just test` - Run tests with coverage
- `make lint` / `just lint` - Run golangci-lint
- `make fmt` / `just fmt` - Format code
- `make release-test` - Test goreleaser release process
