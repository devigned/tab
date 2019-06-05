package opencensus

import (
	"context"
	"encoding/base64"
	"github.com/devigned/tab"
	oct "go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

func init() {
	tab.Register(new(Trace))
}

const (
	propagationKey = "_oc_prop"
)

type (
	// Trace is the implementation of the OpenCensus trace abstraction
	Trace struct{}

	// Span is the implementation of the OpenCensus Span abstraction
	Span struct {
		span *oct.Span
	}
)

// StartSpan starts a new child span of the current span in the context. If
// there is no span in the context, creates a new trace and span.
//
// Returned context contains the newly created span. You can use it to
// propagate the returned span in process.
func (t *Trace) StartSpan(ctx context.Context, operationName string, opts ...interface{}) (context.Context, tab.Spanner) {
	ctx, span := oct.StartSpan(ctx, operationName, toOCOption(opts...)...)
	return ctx, &Span{span: span}
}

// StartSpanWithRemoteParent starts a new child span of the span from the given parent.
//
// If the incoming context contains a parent, it ignores. StartSpanWithRemoteParent is
// preferred for cases where the parent is propagated via an incoming request.
//
// Returned context contains the newly created span. You can use it to
// propagate the returned span in process.
func (t *Trace) StartSpanWithRemoteParent(ctx context.Context, operationName string, carrier tab.Carrier, opts ...interface{}) (context.Context, tab.Spanner) {
	keysValues := carrier.GetKeyValues()
	if val, ok := keysValues[propagationKey]; ok {
		// check if bin and extract
		if bin, ok := val.([]byte); ok {
			if sc, ok := propagation.FromBinary(bin); ok {
				ctx, span := oct.StartSpanWithRemoteParent(ctx, operationName, sc)
				return ctx, &Span{span: span}
			}
		}

		// check if string and expect base64 encoded
		if strVal, ok := val.(string); ok {
			if decoded, err := base64.StdEncoding.DecodeString(strVal); err == nil {
				if sc, ok := propagation.FromBinary(decoded); ok {
					ctx, span := oct.StartSpanWithRemoteParent(ctx, operationName, sc)
					return ctx, &Span{span: span}
				}
			}
		}
	}

	return t.StartSpan(ctx, operationName)
}

// FromContext returns the Span stored in a context, or nil if there isn't one.
func (t *Trace) FromContext(ctx context.Context) tab.Spanner {
	sp := oct.FromContext(ctx)
	return &Span{span: sp}
}

// NewContext returns a new context with the given Span attached.
func (t *Trace) NewContext(ctx context.Context, span tab.Spanner) context.Context {
	if sp, ok := span.InternalSpan().(*oct.Span); ok {
		return oct.NewContext(ctx, sp)
	}
	return ctx
}

// AddAttributes sets attributes in the span.
//
// Existing attributes whose keys appear in the attributes parameter are overwritten.
func (s *Span) AddAttributes(attributes ...tab.Attribute) {
	s.span.AddAttributes(attributesToOCAttributes(attributes...)...)
}

// End ends the span
func (s *Span) End() {
	s.span.End()
}

// Logger returns a trace.Logger for the span
func (s *Span) Logger() tab.Logger {
	return &tab.SpanLogger{Span: s}
}

// Inject propagation key onto the carrier
func (s *Span) Inject(carrier tab.Carrier) error {
	spanBin := propagation.Binary(s.span.SpanContext())
	encodedSpan := base64.StdEncoding.EncodeToString(spanBin)
	carrier.Set(propagationKey, encodedSpan)
	return nil
}

// InternalSpan returns the real implementation of the Span
func (s *Span) InternalSpan() interface{} {
	return s.span
}

func toOCOption(opts ...interface{}) []oct.StartOption {
	var ocStartOptions []oct.StartOption
	for _, opt := range opts {
		if o, ok := opt.(oct.StartOption); ok {
			ocStartOptions = append(ocStartOptions, o)
		}
	}
	return ocStartOptions
}

func attributesToOCAttributes(attributes ...tab.Attribute) []oct.Attribute {
	var ocAttributes []oct.Attribute
	for _, attr := range attributes {
		switch attr.Value.(type) {
		case int64:
			ocAttributes = append(ocAttributes, oct.Int64Attribute(attr.Key, attr.Value.(int64)))
		case string:
			ocAttributes = append(ocAttributes, oct.StringAttribute(attr.Key, attr.Value.(string)))
		case bool:
			ocAttributes = append(ocAttributes, oct.BoolAttribute(attr.Key, attr.Value.(bool)))
		}
	}
	return ocAttributes
}
