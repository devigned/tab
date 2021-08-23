package opentracing

import (
	"testing"

	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
)



func TestKVLogger_Info(t *testing.T) {
	tracer := mocktracer.New()
	span := tracer.StartSpan("test no error")
	l := &KVLogger{Span: &Span{ span }}
	l.Info("info")
	l.Debug("debug")
	span.Finish()

	finishedSpan := tracer.FinishedSpans()[0]
	logs := finishedSpan.Logs()
	assert.Len(t, logs,2)
	assert.Len(t, logs[0].Fields,2)
}