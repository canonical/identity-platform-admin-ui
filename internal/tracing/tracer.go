package tracing

import (
	"context"
	"runtime/debug"

	"github.com/canonical/identity-platform-admin-ui/internal/logging"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
	"go.opentelemetry.io/otel/trace"
)

type Tracer struct {
	tracer trace.Tracer

	logger logging.LoggerInterface
}

func (t *Tracer) init(service string, e sdktrace.SpanExporter) {
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(e),
		sdktrace.WithResource(
			t.buildResource(service),
		),
	)

	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	t.tracer = otel.Tracer(service)
}

func (t *Tracer) gitRevision(settings []debug.BuildSetting) string {
	for _, setting := range settings {
		if setting.Key == "vcs.revision" {
			return setting.Value
		}
	}

	return "n/a"
}

func (t *Tracer) buildResource(service string) *resource.Resource {
	var res *resource.Resource

	res = resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(service),
		semconv.ServiceVersion("n/a"),
	)

	if info, ok := debug.ReadBuildInfo(); ok {
		if service == "" {
			service = info.Path
		}

		res = resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(service),
			attribute.String("git_sha", t.gitRevision(info.Settings)),
			attribute.String("app", info.Main.Path),
		)
	}

	return res
}

func (t *Tracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, spanName, opts...)
}

// basic tracer implementation of trace.Tracer, just adding some extra configuration
func NewTracer(cfg *Config) *Tracer {
	t := new(Tracer)

	t.logger = cfg.Logger

	// if tracing disabled skip the config
	if !cfg.Enabled {
		t.tracer = trace.NewNoopTracerProvider().Tracer("github.com/canonical/identity-platform-admin-ui")

		return t
	}

	var err error
	var exporter sdktrace.SpanExporter

	if cfg.OtelGRPCEndpoint != "" {
		exporter, err = otlptrace.New(
			context.TODO(),
			otlptracegrpc.NewClient(
				otlptracegrpc.WithEndpoint(cfg.OtelGRPCEndpoint),
				otlptracegrpc.WithInsecure(),
			),
		)
	} else if cfg.OtelHTTPEndpoint != "" {
		exporter, err = otlptrace.New(
			context.TODO(),
			otlptracehttp.NewClient(
				otlptracehttp.WithEndpoint(cfg.OtelHTTPEndpoint),
				otlptracehttp.WithInsecure(),
			),
		)
	} else {
		exporter, err = stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
	}

	if err != nil {
		t.logger.Errorf("unable to initialize tracing exporter due: %w", err)
		return nil
	}

	// set tracer provider and propagator properly, this is to ensure all
	// instrumentation library could run well
	t.init("github.com/canonical/identity-platform-admin-ui", exporter)

	return t
}
