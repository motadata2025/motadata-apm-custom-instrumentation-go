# Motadata APM Custom Instrumentation — Go

Add validated, auto-prefixed business attributes to your traces when using **Motadata APM with OpenTelemetry eBPF zero-code auto-instrumentation**.

---

## Requirements

| Requirement | Minimum Version |
|---|---|
| Go | 1.25+ |
| Motadata APM Agent | 8.2.0+ |
| `go.opentelemetry.io/otel` | v1.43.0+ |

---

## How It Works (Go vs Other Languages)

In Java, Python, Node.js, .NET, and PHP, the instrumentation library finds the **current active span** from the SDK's global context and adds attributes to it directly.

**Go with eBPF is different.** eBPF creates spans at kernel level; they are never stored in `context.Context` in your Go app's user space. Calling `trace.SpanFromContext(ctx)` always returns a **no-op span** — attributes set on it are silently dropped.

The correct approach: **create child spans** that inherit the trace context from the eBPF parent. This library provides a thin, validated wrapper around exactly that pattern.

```
eBPF HTTP Span  [auto]
└── CreateUser  [motadata.StartSpan → your custom attrs]
    └── db:CreateUser  [motadata.StartSpan → your custom attrs]
        └── eBPF DB Span  [auto]
```

All spans share the same `TraceID` and appear as a proper tree in the APM UI.

---

## Installation

```bash
go get github.com/motadata2025/motadata-apm-custom-instrumentation-go@latest
```

No other OTel packages need to be imported in your application code — this library handles them internally.

> **Important:** Do **not** call `otel.SetTracerProvider(...)` anywhere in your app. The eBPF agent registers the global `TracerProvider` automatically. Initializing one manually will conflict with the Auto SDK and break span correlation.

---

## Quick Start

```go
import motadata "github.com/motadata2025/motadata-apm-custom-instrumentation-go"

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    ctx, span, _ := motadata.StartSpan(r.Context(), "CreateUser")
    defer span.End()

    _ = span.SetString("user.username", req.Username)
    _ = span.SetInt("user.age", int64(req.Age))
    _ = span.SetBool("user.verified", true)

    result, err := h.repo.CreateUser(ctx, req)
    if err != nil {
        span.RecordError(err)
        return
    }
    // ...
}
```

---

## API Reference

### Span Lifecycle

#### `motadata.StartSpan`

```go
func StartSpan(ctx context.Context, spanName string) (context.Context, *Span, error)
```

Creates a child span linked to the eBPF parent trace. The instrumentation scope is automatically set to `"motadata-go-custom-instrumentation"` — you do not need to provide a service name; the eBPF agent already sets `service.name` as a resource attribute.

| Parameter | Description |
|---|---|
| `ctx` | The current context (e.g. `r.Context()` in HTTP handlers, or the `ctx` returned by a parent `StartSpan` call) |
| `spanName` | Name of the operation (e.g. `"CreateUser"`, `"db:InsertOrder"`) |

Returns `ErrEmptySpanName` if `spanName` is empty or whitespace-only. On error, a no-op span is returned — `defer span.End()` is always safe to call.

```go
ctx, span, err := motadata.StartSpan(r.Context(), "CreateUser")
if err != nil {
    // spanName was empty — handle or log
}
defer span.End()
```

For operations where the span name is a hardcoded constant, ignoring the error is acceptable:

```go
ctx, span, _ := motadata.StartSpan(r.Context(), "CreateUser")
defer span.End()
```

#### `span.End`

```go
func (s *Span) End()
```

Finalizes and exports the span to the backend. Always call immediately after `StartSpan` using `defer`.

#### `span.RecordError`

```go
func (s *Span) RecordError(err error)
```

Records `err` as a span event (with stack trace) and sets the span status to `Error`. No-op if `err` is nil.

---

### Scalar Attribute Setters

All methods:
- Trim and lowercase the key
- Validate that the key contains only alphanumeric characters and dots (`a-z`, `0-9`, `.`)
- Automatically prefix the key with `apm.` (double-prefixing is prevented)
- Return an `error` if the key is invalid or the value is not finite (floats)

#### `span.SetString`

```go
func (s *Span) SetString(key, value string) error
```

```go
_ = span.SetString("user.username", "john_doe")
// Stored as: apm.user.username = "john_doe"
```

#### `span.SetInt`

```go
func (s *Span) SetInt(key string, value int64) error
```

```go
_ = span.SetInt("order.item.count", int64(5))
// Stored as: apm.order.item.count = 5
```

#### `span.SetFloat`

```go
func (s *Span) SetFloat(key string, value float64) error
```

Returns `ErrInvalidFloat` if value is `NaN` or `±Inf`.

```go
_ = span.SetFloat("order.weight.kg", 2.75)
// Stored as: apm.order.weight.kg = 2.75
```

#### `span.SetBool`

```go
func (s *Span) SetBool(key string, value bool) error
```

```go
_ = span.SetBool("order.express", true)
// Stored as: apm.order.express = true
```

---

### Slice / Array Attribute Setters

#### `span.SetStringSlice`

```go
func (s *Span) SetStringSlice(key string, values []string) error
```

```go
_ = span.SetStringSlice("order.tags", []string{"urgent", "vip", "fragile"})
// Stored as: apm.order.tags = ["urgent", "vip", "fragile"]
```

#### `span.SetIntSlice`

```go
func (s *Span) SetIntSlice(key string, values []int64) error
```

```go
_ = span.SetIntSlice("product.ids", []int64{101, 202, 303})
// Stored as: apm.product.ids = [101, 202, 303]
```

#### `span.SetFloatSlice`

```go
func (s *Span) SetFloatSlice(key string, values []float64) error
```

Returns `ErrInvalidFloat` if any element is `NaN` or `±Inf`.

```go
_ = span.SetFloatSlice("item.prices", []float64{9.99, 19.99, 4.50})
// Stored as: apm.item.prices = [9.99, 19.99, 4.50]
```

#### `span.SetBoolSlice`

```go
func (s *Span) SetBoolSlice(key string, values []bool) error
```

```go
_ = span.SetBoolSlice("feature.flags", []bool{true, false, true})
// Stored as: apm.feature.flags = [true, false, true]
```

---

## Usage by Layer

### HTTP Handler

```go
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    ctx, span, _ := motadata.StartSpan(r.Context(), "CreateUser")
    defer span.End()

    _ = span.SetString("http.method", r.Method)
    _ = span.SetString("http.url", r.URL.Path)
    _ = span.SetString("operation", "create_user")
    _ = span.SetString("user.username", req.Username)

    result, err := h.repo.CreateUser(ctx, req)
    if err != nil {
        span.RecordError(err)
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }
    // ...
}
```

### Repository / Data Access

```go
func (r *UserRepo) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    // ctx carries the handler span as parent — child link is automatic
    _, span, _ := motadata.StartSpan(ctx, "db:CreateUser")
    defer span.End()

    _ = span.SetString("db.operation", "INSERT")
    _ = span.SetString("db.table", "users")
    _ = span.SetString("db.param.username", req.Username)

    err := r.db.QueryRowContext(ctx, query, req.Username).Scan(&user.ID)
    if err != nil {
        span.RecordError(err)
        return nil, err
    }
    return &user, nil
}
```

### Service / Business Logic

```go
func ProcessOrder(ctx context.Context, orderID string) error {
    ctx, span, _ := motadata.StartSpan(ctx, "ProcessOrder")
    defer span.End()

    _ = span.SetString("order.id", orderID)
    _ = span.SetBool("order.express", true)
    _ = span.SetInt("order.items", 5)
    _ = span.SetFloat("order.weight.kg", 2.5)
    _ = span.SetStringSlice("order.tags", []string{"urgent", "vip"})

    return nil
}
```

---

## Key Validation Rules

| Rule | Example (input → stored key) |
|---|---|
| Automatically prefixed with `apm.` | `"user.name"` → `apm.user.name` |
| Leading/trailing whitespace trimmed | `"  user.name  "` → `apm.user.name` |
| Converted to lowercase | `"User.Name"` → `apm.user.name` |
| Only `a-z`, `0-9`, `.` allowed | `"user-name"` → `ErrInvalidKey` |
| Empty key rejected | `""` → `ErrEmptyKey` |
| Double prefix prevented | `"apm.user.name"` → `apm.user.name` |
| Float NaN / Inf rejected | `math.NaN()` → `ErrInvalidFloat` |

---

## Error Handling

Attribute errors are **non-fatal**. The span continues normally; only that one attribute is skipped.

```go
if err := span.SetString("user.email", email); err != nil {
    log.Printf("custom attribute skipped: %v", err)
}
```

Sentinel errors:

| Error | Cause |
|---|---|
| `motadata.ErrEmptyKey` | Attribute key is empty or whitespace-only |
| `motadata.ErrInvalidKey` | Attribute key contains characters other than `a-z`, `0-9`, `.` |
| `motadata.ErrInvalidFloat` | Float value is `NaN` or `±Inf` |
| `motadata.ErrEmptySpanName` | `spanName` passed to `StartSpan` is empty or whitespace-only |

---

## Best Practices

- **Use hierarchical dot-separated keys** — `"order.customer.tier"`, `"payment.method"` — for organized grouping in the APM UI
- **Always pass `ctx` downstream** — the child span chain only works if `ctx` flows through every function call
- **Always `defer span.End()`** — a span that is never ended is never exported
- **Keep under 10 attributes per span** — the eBPF buffer has a fixed limit per span; split across child spans if you need more
- **Never let attribute errors break business logic** — always use `_ =` or check and log; attribute failures are non-fatal by design
- **Be consistent with key naming** — use the same key names across services so attributes are filterable across traces in the APM UI

---

## Example Application

A complete working example is available in the [`example/`](./example/) directory, demonstrating HTTP handler, repository, and service layer instrumentation.

---

## Support

For questions or issues: **engg@motadata.com**

---

## License

Copyright © 2026 Motadata. All rights reserved. Proprietary and confidential.
