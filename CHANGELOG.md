# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.0.0] - 2026-04-20

### Added
- `StartSpan(ctx, serviceName, spanName)` — creates a child span linked to the eBPF parent trace via the global `TracerProvider`
- `(*Span).End()` — finalizes and exports the span to the backend
- `(*Span).RecordError(err)` — records an error as a span event with stack trace and sets span status to `Error`
- Scalar attribute setters: `SetString`, `SetInt`, `SetFloat`, `SetBool`
- Slice attribute setters: `SetStringSlice`, `SetIntSlice`, `SetFloatSlice`, `SetBoolSlice`
- Automatic key validation: trim, lowercase, alphanumeric + dots only, auto-prefix with `apm.`
- Double-prefix prevention: `"apm.user.name"` input is stored as `apm.user.name`, not `apm.apm.user.name`
- Sentinel errors: `ErrEmptyKey`, `ErrInvalidKey`, `ErrInvalidFloat`
- Rejection of `NaN` and `±Inf` float values
- Full example application in `example/` demonstrating HTTP handler, repository, and service layers
- Compatible with `go.opentelemetry.io/otel v1.17.0` (Go 1.19+)
