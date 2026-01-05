package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracerName = "barbod-gateway"
)

func OpenTelemetryMiddleware(serviceName string) gin.HandlerFunc {
	tracer := otel.Tracer(tracerName)
	meter := otel.Meter(tracerName)

	requestCounter, _ := meter.Int64Counter("http.server.request_count",
		metric.WithDescription("Total number of HTTP requests"),
	)
	requestDuration, _ := meter.Float64Histogram("http.server.duration",
		metric.WithDescription("Duration of HTTP requests"),
		metric.WithUnit("ms"),
	)

	return func(c *gin.Context) {
		start := time.Now()
		ctx := otel.GetTextMapPropagator().Extract(
			c.Request.Context(),
			propagation.HeaderCarrier(c.Request.Header),
		)

		spanName := fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPMethod(c.Request.Method),
				semconv.HTTPRoute(c.FullPath()),
				semconv.HTTPTarget(c.Request.URL.Path),
				semconv.HTTPScheme(c.Request.URL.Scheme),
				semconv.NetHostName(c.Request.Host),
				semconv.UserAgentOriginal(c.Request.UserAgent()),
				attribute.String("client.address", c.ClientIP()),
				attribute.String("http.request_id", c.GetHeader("X-Request-ID")),
			),
		)
		defer span.End()

		c.Request = c.Request.WithContext(ctx)

		c.Next()

		status := c.Writer.Status()
		span.SetAttributes(
			semconv.HTTPStatusCode(status),
			attribute.Int("http.response_size", c.Writer.Size()),
		)

		duration := time.Since(start).Seconds() * 1000
		attrs := []attribute.KeyValue{
			semconv.HTTPMethod(c.Request.Method),
			semconv.HTTPRoute(c.FullPath()),
			semconv.HTTPStatusCode(status),
		}

		requestCounter.Add(c.Request.Context(), 1, metric.WithAttributes(attrs...))
		requestDuration.Record(c.Request.Context(), duration, metric.WithAttributes(attrs...))

		if status >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", status))
			if len(c.Errors) > 0 {
				span.RecordError(c.Errors.Last())
				span.SetAttributes(attribute.String("error.message", c.Errors.Last().Error()))
			}
		} else {
			span.SetStatus(codes.Ok, "")
		}
	}
}

func GetSpanFromContext(c *gin.Context) trace.Span {
	return trace.SpanFromContext(c.Request.Context())
}

func AddSpanEvent(c *gin.Context, name string, attributes ...attribute.KeyValue) {
	span := GetSpanFromContext(c)
	span.AddEvent(name, trace.WithAttributes(attributes...))
}

func AddSpanAttributes(c *gin.Context, attributes ...attribute.KeyValue) {
	span := GetSpanFromContext(c)
	span.SetAttributes(attributes...)
}
