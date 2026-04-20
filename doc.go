// Package motadata provides custom attribute instrumentation for Go applications
// running under Motadata APM with OpenTelemetry eBPF zero-code auto-instrumentation.
//
// # Why this package exists
//
// eBPF auto-instrumentation creates spans at kernel level. Those spans are never
// stored in context.Context in user space, so trace.SpanFromContext always returns
// a no-op span — any attributes set on it are silently dropped.
//
// This package solves that by using autosdk.TracerProvider() directly — the Auto SDK
// bridges user-space spans to the eBPF parent trace automatically, then providing
// validated, auto-prefixed attribute setters on those spans.
//
// # Basic usage
//
//	import motadata "github.com/motadata2025/motadata-apm-custom-instrumentation-go"
//
//	func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
//	    ctx, span, _ := motadata.StartSpan(r.Context(), "CreateUser")
//	    defer span.End()
//
//	    _ = span.SetString("user.name", req.Username)  // stored as apm.user.name
//	    _ = span.SetInt("user.age", int64(req.Age))     // stored as apm.user.age
//
//	    result, err := h.repo.CreateUser(ctx, req)
//	    if err != nil {
//	        span.RecordError(err)
//	        return
//	    }
//	}
//
// # Key rules
//
// All attribute keys are trimmed, lowercased, validated (alphanumeric + dots only),
// and automatically prefixed with "apm.". Passing "apm." already in the key is safe —
// double-prefixing is prevented.
//
// Do NOT call otel.SetTracerProvider anywhere in your application. This package
// uses autosdk.TracerProvider() directly — the eBPF agent hooks into it automatically.
// Initializing a global TracerProvider manually conflicts with the Auto SDK and breaks
// span correlation.
package motadata
