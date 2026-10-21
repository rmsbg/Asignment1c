# Repository Guidelines

## Project Structure & Module Organization

The root `README.md` defines BikeRental: adapt the existing ski rental application for bicycles, hourly/daily rentals, bike categories, optional equipment/insurance, and bike photos. Reuse existing functionality wherever practical and retain its technology stack.

The working Go module and Git repository are in `SkiRental/`:
- `cmd/server/`: application entry point.
- `internal/domain/`: models, validation, and pricing rules.
- `internal/store/`: thread-safe, in-memory storage and rental operations.
- `internal/handler/`: HTTP handlers and API tests.
- `web/`: embedded Go HTML templates and plain JavaScript/CSS assets.
- `specs/`: requirements, acceptance records, and screenshots.
- `testdata/`: synthetic fixtures. Go tests sit beside implementation files.

Read `SkiRental/AGENTS.md` before editing that subtree. Its initial project-status description is outdated; inspect current code and README for implemented functionality.

## Build, Test, and Development Commands

Use Go 1.27.3 or newer. Run these commands from `SkiRental/` (`cd SkiRental`):

```sh
go run ./cmd/server       # Start at http://127.0.0.1:8080
go run ./cmd/server -demo # Start with synthetic sample data
go build ./...           # Compile all packages
go test ./...            # Run automated tests
go test -race ./...      # Check concurrent access; requires a C compiler
go vet ./...             # Run static analysis
go fmt ./...             # Format Go code
```

There is no frontend build step or external database. Restart the server after changing embedded assets; restarting discards stored data.

## Coding Style & Naming Conventions

Use `gofmt` tab indentation, lowercase Go package names, exported `PascalCase`, and unexported `camelCase` identifiers. Follow nearby JavaScript/CSS conventions. Keep business rules in domain/store packages and HTTP concerns in handlers. Prefer the standard library; avoid unnecessary dependencies. Wrap propagated errors with context using `%w`.

## Testing Guidelines

Use Go's `testing` package, `*_test.go` filenames, and `TestXxx` functions. Cover pricing, availability conflicts, validation, and HTTP workflows. No coverage percentage is required. Run race tests for storage changes and manually verify affected browser flows, including mobile layout and keyboard use. Record acceptance evidence in `specs/`.

## Commit & Pull Request Guidelines

Existing commits use German phase-based subjects, such as `Phase 5: Verleihtermine und Tarifverwaltung abschließen`. Document explicit customer acceptance before creating phase-completion commits. PRs should explain changed behavior, link requirements or issues, report validation, and include screenshots for UI changes.

## Security & Configuration

Use synthetic customer data only. Never commit secrets or real identity documents, and keep identity details out of logs and overview screens. Preserve loopback-only operation until authentication and deployment requirements are defined.
