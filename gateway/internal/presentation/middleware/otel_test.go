package middleware

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func installTestTracer(t *testing.T) (*tracetest.SpanRecorder, *sdktrace.TracerProvider, func()) {
	t.Helper()
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	cleanup := func() { _ = tp.Shutdown(context.Background()) }
	return sr, tp, cleanup
}

func TestOpenTelemetryMiddleware_Success(t *testing.T) {
	sr, _, cleanup := installTestTracer(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(OpenTelemetryMiddleware("testservice"))
	r.GET("/hello/:id", func(c *gin.Context) {
		AddSpanEvent(c, "test-event", attribute.String("k", "v"))
		AddSpanAttributes(c, attribute.String("extra.attr", "val"))
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/hello/1", nil)
	req.Header.Set("X-Request-ID", "req-1")
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	spans := sr.Ended()
	assert.Equal(t, 1, len(spans))
	s := spans[0]
	// verify status ok
	assert.Equal(t, int(codes.Ok), int(s.Status().Code))
	// expect attribute we added
	attrs := s.Attributes()
	// semconv HTTP route + method exist
	assert.Contains(t, attrs, semconv.HTTPRoute("/hello/:id"))
	assert.Contains(t, attrs, semconv.HTTPMethod("GET"))
	// our custom attribute
	assert.Contains(t, attrs, attribute.String("extra.attr", "val"))
	// event added
	events := s.Events()
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "test-event", events[0].Name)
	// request id attribute present
	assert.Contains(t, attrs, attribute.String("http.request_id", "req-1"))
}

func TestOpenTelemetryMiddleware_ErrorStatusAndRecordsError(t *testing.T) {
	sr, _, cleanup := installTestTracer(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(OpenTelemetryMiddleware("testservice"))
	r.GET("/err", func(c *gin.Context) {
		c.Error(errors.New("boom"))
		c.String(500, "err")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/err", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)

	spans := sr.Ended()
	assert.Equal(t, 1, len(spans))
	s := spans[0]
	assert.Equal(t, int(codes.Error), int(s.Status().Code)) // should be error
	// error message attribute should be set
	assert.Contains(t, s.Attributes(), attribute.String("error.message", "boom"))
}

func TestGetSpanFromContextAndHelpers(t *testing.T) {
	sr, _, cleanup := installTestTracer(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(OpenTelemetryMiddleware("srv"))
	r.GET("/helpers", func(c *gin.Context) {
		AddSpanEvent(c, "e1", attribute.String("a1", "v1"))
		AddSpanAttributes(c, attribute.String("a2", "v2"))
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/helpers", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	spans := sr.Ended()
	assert.Equal(t, 1, len(spans))
	s := spans[0]
	// event was added and attribute was set
	assert.Equal(t, 1, len(s.Events()))
	assert.Contains(t, s.Attributes(), attribute.String("a2", "v2"))
}
