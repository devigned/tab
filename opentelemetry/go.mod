module github.com/devigned/tab/opentelemetry

go 1.16

require (
	github.com/devigned/tab v0.0.1
	github.com/opentracing/opentracing-go v1.1.0
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel v1.4.1 // indirect
	go.opentelemetry.io/otel/trace v1.4.1 // indirect
)

replace github.com/devigned/tab => ../.
