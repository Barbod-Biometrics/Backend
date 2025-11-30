package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/bootstrap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)


type Telemetry struct {	
	tracerProvider *trace.TracerProvider
	meterProvider  *metric.MeterProvider
	resource       *resource.Resource
}

func InitTelemetry(ctx context.Context, config *bootstrap.Telemetry) (*Telemetry, error) {
	if !config.Enabled {
		return &Telemetry{}, nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.ServiceVersionKey.String(config.ServiceVersion),
			semconv.DeploymentEnvironmentKey.String(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tracerProvider, err := initTracerProvider(ctx, config.OTLPEndpoint, res)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tracer provider: %w", err)
	}

	meterProvider, err := initMeterProvider(ctx, config.OTLPEndpoint, res)
	if err != nil {
		tracerProvider.Shutdown(ctx)
		return nil, fmt.Errorf("failed to initialize meter provider: %w", err)
	}

	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(meterProvider)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Telemetry{
		tracerProvider: tracerProvider,
		meterProvider:  meterProvider,
		resource:       res,
	}, nil
}

func initTracerProvider(ctx context.Context, endpoint string, res *resource.Resource) (*trace.TracerProvider, error) {
	traceExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(extractEndpoint(endpoint)),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			trace.WithBatchTimeout(5*time.Second),
		),
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()),
	)

	return tracerProvider, nil
}

func initMeterProvider(ctx context.Context, endpoint string, res *resource.Resource) (*metric.MeterProvider, error) {
	metricExporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(extractEndpoint(endpoint)),
		otlpmetrichttp.WithInsecure(), 
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			metric.WithInterval(30*time.Second),
		)),
		metric.WithResource(res),
	)

	return meterProvider, nil
}


func (t *Telemetry) Shutdown(ctx context.Context) error {
	if t.tracerProvider == nil && t.meterProvider == nil {
		return nil
	}

	var err error
	if t.tracerProvider != nil {
		if shutdownErr := t.tracerProvider.Shutdown(ctx); shutdownErr != nil {
			err = fmt.Errorf("failed to shutdown tracer provider: %w", shutdownErr)
		}
	}

	if t.meterProvider != nil {
		if shutdownErr := t.meterProvider.Shutdown(ctx); shutdownErr != nil {
			if err != nil {
				err = fmt.Errorf("%v; failed to shutdown meter provider: %w", err, shutdownErr)
			} else {
				err = fmt.Errorf("failed to shutdown meter provider: %w", shutdownErr)
			}
		}
	}

	return err
}


func extractEndpoint(endpoint string) string {
	// Remove http:// or https:// prefix if present
	if len(endpoint) > 7 && endpoint[:7] == "http://" {
		return endpoint[7:]
	}
	if len(endpoint) > 8 && endpoint[:8] == "https://" {
		return endpoint[8:]
	}
	return endpoint
}