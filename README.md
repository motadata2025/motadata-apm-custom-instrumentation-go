# Motadata APM Custom Instrumentation — Go

Add custom business attributes to your traces when using **Motadata APM with OpenTelemetry eBPF zero-code auto-instrumentation**.

---

## Requirements

| Requirement | Minimum Version |
|---|---|
| Go | 1.19+ |
| Motadata Agent | 8.1.0+ |
| `go.opentelemetry.io/otel` | v1.17.0+ |

> **OTel version note:** This package ships with `v1.17.0` (minimum Go 1.19). If eBPF span correlation does not work in your environment, upgrade using the fallback ladder: `v1.21.0` → `v1.28.0` → `v1.36.0`. Only the `require` line in your `go.mod` changes — no code changes needed.

---

## How It Works (Go vs Other Languages)

In Java, Python, Node.js, .NET, and PHP, the instrumentation library finds the **current active span** from the SDK's global context and adds attributes to it.

**Go with eBPF is different.** eBPF creates spans at kernel level; they are never stored in `context.Context` in your Go app's user space. Calling `trace.SpanFromContext(ctx)` always returns a **no-op span** — attributes set on it are silently dropped.

The correct approach: **create child spans** that inherit the trace context from the eBPF parent. This library provides a thin, validated wrapper around exactly that pattern.

```
eBPF HTTP Span  [auto]
└── CreateUser  [motadata.StartSpan → your custom attrs]
    └── db:CreateUser  [motadata.StartSpan → your custom attrs]
        └── eBPF DB Span  [auto]
```

All spans share the same `TraceID` and appear as a proper tree.

---

## Installation

```bash
go get github.com/motadata2025/motadata-apm-custom-instrumentation-go@latest
```

No other OTel packages need to be imported in your application code — this library handles them.

> **Important:** Do **not** call `otel.SetTracerProvider(...)` anywhere in your app. The eBPF agent registers the global `TracerProvider` automatically. Initializing one manually will conflict with the Auto SDK and break span correlation.

---

## Quick Start

```go
import motadata "github.com/motadata2025/motadata-apm-custom-instrumentation-go"

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    ctx, span := motadata.StartSpan(r.Context(), "user-service", "CreateUser")
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
func StartSpan(ctx context.Context, serviceName, spanName string) (context.Context, *Span)
```

Creates a child span linked to the eBPF parent trace.

| Parameter | Description |
|---|---|
| `ctx` | The current context (e.g., `r.Context()` in HTTP handlers, or the `ctx` passed from a parent span) |
| `serviceName` | Name of your service (e.g., `"user-service"`, `"order-service"`) |
| `spanName` | Name of the operation (e.g., `"CreateUser"`, `"db:InsertOrder"`) |

Always call `defer span.End()` immediately after `StartSpan`:

```go
ctx, span := motadata.StartSpan(r.Context(), "my-service", "MyOperation")
defer span.End()
```

#### `span.End`

```go
func (s *Span) End()
```

Finalizes and exports the span to the backend.

#### `span.RecordError`

```go
func (s *Span) RecordError(err error)
```

Records `err` as a span event (with stack trace) and sets span status to `Error`. No-op if `err` is nil.

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

Returns `ErrInvalidFloat` if any value is `NaN` or `±Inf`.

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
    ctx, span := motadata.StartSpan(r.Context(), "user-service", "CreateUser")
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
    _, span := motadata.StartSpan(ctx, "user-service", "db:CreateUser")
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
    ctx, span := motadata.StartSpan(ctx, "order-service", "ProcessOrder")
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

```go
if err := span.SetString("user.email", email); err != nil {
    log.Printf("custom attribute error: %v", err)
    // span is still active — other attributes and the span itself are unaffected
}
```

Attribute errors are non-fatal. The span continues normally; only that one attribute is not recorded.

Sentinel errors:

| Error | Cause |
|---|---|
| `motadata.ErrEmptyKey` | Key is empty or whitespace-only |
| `motadata.ErrInvalidKey` | Key contains characters other than `a-z`, `0-9`, `.` |
| `motadata.ErrInvalidFloat` | Float value is `NaN` or `±Inf` |

---

## Best Practices

- **Use hierarchical keys** — `"user.id"`, `"db.table"`, `"order.status"` — for organized attribute grouping in the APM UI
- **Always pass `ctx` downstream** — the child span chain (`StartSpan(ctx, ...)`) only works if `ctx` flows through every function call
- **Always `defer span.End()`** — a span that is never ended is never exported
- **Wrap in try/catch equivalent** — attribute errors are safe to log and ignore; never let them break business logic
- **Be consistent with key naming** — use the same key names across services so attributes are filterable across traces

---

## Example Application

A complete working example is available in the [`example/`](./example/) directory.

---

## Support

For questions or issues: **engg@motadata.com**

---

## License

Copyright © 2026 Motadata. All rights reserved. Proprietary and confidential.
