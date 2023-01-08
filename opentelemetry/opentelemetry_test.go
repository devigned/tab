package opentelemetry

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestKVLogger_Info(t *testing.T) {
	tracer := trace.NewNoopTracerProvider().Tracer("")
	_, span := tracer.Start(context.Background(), "test no error")
	l := &KVLogger{Span: &Span{span}}
	l.Info("info")
	l.Debug("debug")
	span.End()

	// finishedSpan := tracer.FinishedSpans()[0]
	// logs := finishedSpan.Logs()
	// assert.Len(t, logs, 2)
	// assert.Len(t, logs[0].Fields, 2)
}
