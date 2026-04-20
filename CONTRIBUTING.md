# Contributing to Motadata APM Custom Instrumentation — Go

Thank you for your interest in contributing. Please read these guidelines before submitting changes.

---

## Setup

```bash
git clone https://github.com/motadata2025/motadata-apm-custom-instrumentation-go.git
cd motadata-apm-custom-instrumentation-go
go mod download
```

Run tests:

```bash
go test ./...
```

---

## Code Quality Requirements

- Follow standard [Go formatting](https://go.dev/doc/effective_go) — run `gofmt -w .` before committing
- All exported types and functions must have a doc comment
- Use standard `error` return values — no custom error types (use sentinel errors like `ErrEmptyKey`)
- No third-party dependencies beyond `go.opentelemetry.io/auto/sdk`, `go.opentelemetry.io/otel`, and `go.opentelemetry.io/otel/trace`
- All new behavior must be covered by tests in `*_test.go` files

---

## Workflow

1. Create a feature or fix branch from `main`
2. Make your changes with tests
3. Run `go test ./...` and `go vet ./...` — both must pass
4. Update `README.md` if any public API changes
5. Add an entry to `CHANGELOG.md` under an `[Unreleased]` section
6. Submit a pull request with a clear description of the change

---

## OTel Version Policy

This package targets `go.opentelemetry.io/otel v1.43.0` and Go 1.25+. Do not introduce APIs or behaviors that require a higher OTel version without updating `README.md` and bumping `go.mod` accordingly.

---

## Release

1. Update the version in `go.mod` (if applicable)
2. Move `[Unreleased]` entries in `CHANGELOG.md` to a new versioned section with today's date
3. Tag the release: `git tag v1.x.x && git push origin v1.x.x`

---

## Contact

Motadata Engineering Team — engg@motadata.com
