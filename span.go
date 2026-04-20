package motadata

import (
	"context"
	"math"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Span wraps an OpenTelemetry span with validated, auto-prefixed attribute setters.
type Span struct {
	span trace.Span
}

// StartSpan creates a child span under the given context using the Motadata eBPF Auto SDK.
// serviceName identifies the instrumented service; spanName names the operation.
// Always call span.End() (typically via defer) immediately after StartSpan.
//
// Do NOT call otel.SetTracerProvider anywhere in your app — the eBPF agent
// registers the provider globally. This function uses it automatically.
func StartSpan(ctx context.Context, serviceName, spanName string) (context.Context, *Span) {
	tracer := otel.Tracer(serviceName)
	ctx, s := tracer.Start(ctx, spanName)
	return ctx, &Span{span: s}
}

// End finalizes and exports the span to the backend.
// Call immediately after StartSpan using defer:
//
//	ctx, span := motadata.StartSpan(r.Context(), "svc", "op")
//	defer span.End()
func (s *Span) End() {
	s.span.End()
}

// RecordError records err as a span event with a stack trace and sets
// the span status to Error. No-op if err is nil.
func (s *Span) RecordError(err error) {
	if err == nil {
		return
	}
	s.span.RecordError(err)
	s.span.SetStatus(codes.Error, err.Error())
}

// SetString sets a string attribute on the span.
// key is trimmed, lowercased, validated, and automatically prefixed with "apm.".
func (s *Span) SetString(key, value string) error {
	k, err := validateKey(key)
	if err != nil {
		return err
	}
	s.span.SetAttributes(attribute.String(k, value))
	return nil
}

// SetInt sets an int64 attribute on the span.
// key is trimmed, lowercased, validated, and automatically prefixed with "apm.".
func (s *Span) SetInt(key string, value int64) error {
	k, err := validateKey(key)
	if err != nil {
		return err
	}
	s.span.SetAttributes(attribute.Int64(k, value))
	return nil
}

// SetFloat sets a float64 attribute on the span.
// key is trimmed, lowercased, validated, and automatically prefixed with "apm.".
// Returns ErrInvalidFloat if value is NaN or infinite.
func (s *Span) SetFloat(key string, value float64) error {
	k, err := validateKey(key)
	if err != nil {
		return err
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return ErrInvalidFloat
	}
	s.span.SetAttributes(attribute.Float64(k, value))
	return nil
}

// SetBool sets a bool attribute on the span.
// key is trimmed, lowercased, validated, and automatically prefixed with "apm.".
func (s *Span) SetBool(key string, value bool) error {
	k, err := validateKey(key)
	if err != nil {
		return err
	}
	s.span.SetAttributes(attribute.Bool(k, value))
	return nil
}

// SetStringSlice sets a []string attribute on the span.
// key is trimmed, lowercased, validated, and automatically prefixed with "apm.".
func (s *Span) SetStringSlice(key string, values []string) error {
	k, err := validateKey(key)
	if err != nil {
		return err
	}
	s.span.SetAttributes(attribute.StringSlice(k, values))
	return nil
}

// SetIntSlice sets a []int64 attribute on the span.
// key is trimmed, lowercased, validated, and automatically prefixed with "apm.".
func (s *Span) SetIntSlice(key string, values []int64) error {
	k, err := validateKey(key)
	if err != nil {
		return err
	}
	s.span.SetAttributes(attribute.Int64Slice(k, values))
	return nil
}

// SetFloatSlice sets a []float64 attribute on the span.
// key is trimmed, lowercased, validated, and automatically prefixed with "apm.".
// Returns ErrInvalidFloat if any value is NaN or infinite.
func (s *Span) SetFloatSlice(key string, values []float64) error {
	k, err := validateKey(key)
	if err != nil {
		return err
	}
	for _, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return ErrInvalidFloat
		}
	}
	s.span.SetAttributes(attribute.Float64Slice(k, values))
	return nil
}

// SetBoolSlice sets a []bool attribute on the span.
// key is trimmed, lowercased, validated, and automatically prefixed with "apm.".
func (s *Span) SetBoolSlice(key string, values []bool) error {
	k, err := validateKey(key)
	if err != nil {
		return err
	}
	s.span.SetAttributes(attribute.BoolSlice(k, values))
	return nil
}
