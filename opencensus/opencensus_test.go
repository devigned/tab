package opencensus

import (
	"context"
	"fmt"
	"testing"

	oct "go.opencensus.io/trace"

	"github.com/stretchr/testify/assert"
)

type testExporter struct {
	SpanData  *oct.SpanData
}

func (t *testExporter) ExportSpan(s *oct.SpanData) {
	t.SpanData = s
}

var _ oct.Exporter = &testExporter{}

func TestAnnotationLoggerNoError(t *testing.T) {
	exp := &testExporter{}
	oct.RegisterExporter(exp)
	_, s := oct.StartSpan(context.TODO(), "test no error", oct.WithSampler(oct.AlwaysSample()))

	l := &AnnotationLogger{Span: &Span{
		span : s,
	}}
	l.Info("info")
	l.Debug("debug")
	s.End()
	assert.Len(t, exp.SpanData.Attributes,0)
	assert.Equal(t, "info", exp.SpanData.Annotations[0].Message)
	assert.Equal(t, "info", exp.SpanData.Annotations[0].Attributes["level"])
	assert.Equal(t, "debug", exp.SpanData.Annotations[1].Message)
	assert.Equal(t, "debug", exp.SpanData.Annotations[1].Attributes["level"])
}

func TestAnnotationLoggerWithError(t *testing.T) {
	exp := &testExporter{}
	oct.RegisterExporter(exp)
	_, s := oct.StartSpan(context.TODO(), "test no error", oct.WithSampler(oct.AlwaysSample()))

	l := &AnnotationLogger{Span: &Span{
		span : s,
	}}
	l.Error(fmt.Errorf("error"))
	s.End()
	assert.Len(t, exp.SpanData.Attributes,1)
	assert.Equal(t,true, exp.SpanData.Attributes["error"])
	assert.Equal(t, "error", exp.SpanData.Annotations[0].Message)
	assert.Equal(t, "error", exp.SpanData.Annotations[0].Attributes["level"])
}

func TestAnnotationLoggerWithFatal(t *testing.T) {
	exp := &testExporter{}
	oct.RegisterExporter(exp)
	_, s := oct.StartSpan(context.TODO(), "test no error", oct.WithSampler(oct.AlwaysSample()))

	l := &AnnotationLogger{Span: &Span{
		span : s,
	}}
	l.Fatal("fatal")
	s.End()
	assert.Len(t, exp.SpanData.Attributes,1)
	assert.Equal(t,true, exp.SpanData.Attributes["error"])
	assert.Equal(t, "fatal", exp.SpanData.Annotations[0].Message)
	assert.Equal(t, "fatal", exp.SpanData.Annotations[0].Attributes["level"])
}
