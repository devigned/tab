package opentelemetry

import (
	"context"
	"fmt"

	"github.com/devigned/tab"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var Tracer trace.Tracer

var _ propagation.TextMapCarrier = (*carrierAdapter)(nil)

type carrierAdapter struct {
	carrier tab.Carrier
}

// Get returns the value associated with the passed key.
func (ca carrierAdapter) Get(key string) string {
	values := ca.carrier.GetKeyValues()
	for k, v := range values {
		if k == key {
			if sv, ok := v.(string); ok {
				return sv
			}
			return ""
		}
	}
	return ""
}

// Set stores the key-value pair.
func (ca carrierAdapter) Set(key string, value string) {
	ca.carrier.Set(key, value)
}

// Keys lists the keys stored in this carrier.
func (ca carrierAdapter) Keys() []string {
	kv := ca.carrier.GetKeyValues()
	keys := []string{}
	for k := range kv {
		keys = append(keys, k)
	}
	return keys
}

func init() {
	tab.Register(new(Trace))
}

type (
	// Trace is the implementation of the OpenTelemetry trace abstraction
	Trace struct{}

	// Span is the implementation of the OpenTelemetry Span abstraction
	Span struct {
		span trace.Span
	}
)

// StartSpan starts and returns a Span with `operationName`, using
// any Span found within `ctx` as a ChildOfRef. If no such parent could be
// found, StartSpanFromContext creates a root (parentless) Span.
func (t *Trace) StartSpan(ctx context.Context, operationName string, opts ...interface{}) (context.Context, tab.Spanner) {
	s := &Span{}
	if Tracer != nil {
		ctx, s.span = Tracer.Start(ctx, operationName, toOTOption(opts...)...)
	}
	return ctx, s
}

func extract(reader propagation.TextMapCarrier) (trace.SpanContext, error) {
	ctx := context.Background()
	ctx = otel.GetTextMapPropagator().Extract(ctx, reader)
	span := trace.SpanFromContext(ctx)
	return span.SpanContext(), nil
}

// StartSpanWithRemoteParent starts and returns a Span with `operationName`, using
// reference span as FollowsFrom
func (t *Trace) StartSpanWithRemoteParent(ctx context.Context, operationName string, carrier tab.Carrier, opts ...interface{}) (context.Context, tab.Spanner) {
	sc, err := extract(&carrierAdapter{carrier})
	if err != nil {
		return t.StartSpan(ctx, operationName)
	}

	s := &Span{}
	if Tracer != nil {
		ctx = trace.ContextWithRemoteSpanContext(ctx, sc)
		ctx, s.span = Tracer.Start(ctx, operationName, toOTOption(opts...)...)
		ctx = trace.ContextWithSpan(ctx, s.span)
	}
	return ctx, s
}

// FromContext returns the `Span` previously associated with `ctx`, or
// `nil` if no such `Span` could be found.
func (t *Trace) FromContext(ctx context.Context) tab.Spanner {
	sp := trace.SpanFromContext(ctx)
	return &Span{span: sp}
}

// NewContext returns a new context with the given Span attached.
func (t *Trace) NewContext(ctx context.Context, span tab.Spanner) context.Context {
	if sp, ok := span.InternalSpan().(trace.Span); ok {
		return trace.ContextWithSpan(ctx, sp)
	}
	return ctx
}

// AddAttributes a tags to the span.
//
// If there is a pre-existing tag set for `key`, it is overwritten.
func (s *Span) AddAttributes(attributes ...tab.Attribute) {
	for _, attr := range attributes {
		s.span.SetAttributes(
			//TODO check type to use the correct attribute
			attribute.String(attr.Key, fmt.Sprintf("%v", attr.Value)),
		)
	}
}

// End sets the end timestamp and finalizes Span state.
//
// With the exception of calls to Context() (which are always allowed),
// Finish() must be the last call made to any span instance, and to do
// otherwise leads to undefined behavior.
func (s *Span) End() {
	s.span.End()
}

// Logger returns a trace.Logger for the span
func (s *Span) Logger() tab.Logger {
	return &KVLogger{Span: s}
}

func inject(s *Span, writer propagation.TextMapCarrier) error {
	ctx := trace.ContextWithSpan(context.Background(), s.span)
	otel.GetTextMapPropagator().Inject(ctx, writer)
	return nil
}

// Inject span context into carrier
func (s *Span) Inject(carrier tab.Carrier) error {
	return inject(s, &carrierAdapter{carrier})
}

// InternalSpan returns the real implementation of the Span
func (s *Span) InternalSpan() interface{} {
	return s.span
}

func toOTOption(opts ...interface{}) []trace.SpanStartOption {
	var ocStartOptions []trace.SpanStartOption
	for _, opt := range opts {
		if o, ok := opt.(trace.SpanStartOption); ok {
			ocStartOptions = append(ocStartOptions, o)
		}
	}
	return ocStartOptions
}

type KVLogger struct {
	Span *Span
}

func (a *KVLogger) Info(msg string, attributes ...tab.Attribute) {
	a.logToAnnotation("info", msg, attributes...)
}

func (a *KVLogger) Error(err error, attributes ...tab.Attribute) {
	a.Span.AddAttributes(tab.BoolAttribute("error", true))
	a.logToAnnotation("error", err.Error(), attributes...)
}

func (a *KVLogger) Fatal(msg string, attributes ...tab.Attribute) {
	a.Span.AddAttributes(tab.BoolAttribute("error", true))
	a.logToAnnotation("fatal", msg, attributes...)
}

func (a *KVLogger) Debug(msg string, attributes ...tab.Attribute) {
	a.logToAnnotation("debug", msg, attributes...)
}

func (a *KVLogger) logToAnnotation(level string, msg string, attributes ...tab.Attribute) {
	attrs := append(attributes,
		tab.StringAttribute("level", level),
		tab.StringAttribute("msg", msg),
	)
	a.Span.AddAttributes(attrs...)
}
