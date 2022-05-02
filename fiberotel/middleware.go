package fiberotel

import (
	"bytes"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"text/template"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

const LocalsCtxKey = "otel-ctx"

var Tracer = otel.Tracer("fiber_otel-router")

// New creates a new middleware handler
func New(config ...Config) fiber.Handler {
	// Set default config
	cfg := configDefault(config...)

	spanTmpl := template.Must(template.New("span").Parse(cfg.SpanName))

	// Return new handler
	return func(c *fiber.Ctx) error {
		// concat all span options, dynamic and static
		spanOptions := concatSpanOptions(
			[]trace.SpanStartOption{
				trace.WithAttributes(semconv.HTTPMethodKey.String(c.Method())),
				trace.WithAttributes(semconv.HTTPTargetKey.String(string(c.Request().RequestURI()))),
				trace.WithAttributes(semconv.HTTPRouteKey.String(c.Route().Path)),
				trace.WithAttributes(semconv.HTTPURLKey.String(c.OriginalURL())),
				trace.WithAttributes(semconv.NetHostIPKey.String(c.IP())),
				trace.WithAttributes(semconv.HTTPUserAgentKey.String(string(c.Request().Header.UserAgent()))),
				trace.WithAttributes(semconv.HTTPRequestContentLengthKey.Int(c.Request().Header.ContentLength())),
				trace.WithAttributes(semconv.HTTPSchemeKey.String(c.Protocol())),
				trace.WithAttributes(semconv.NetTransportTCP),
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(attribute.StringSlice("http.x-forwarded-for", c.IPs())),
				trace.WithAttributes(attribute.String("http.x_forwarded_for", c.GetReqHeaders()["x-forwarded-for"])),
			},
			cfg.TracerStartAttributes,
		)

		spanName := new(bytes.Buffer)
		err := spanTmpl.Execute(spanName, c)
		if err != nil {
			return fmt.Errorf("cannot execute span name template: %w", err)
		}

		otelCtx, span := Tracer.Start(
			c.UserContext(),
			spanName.String(),
			spanOptions...,
		)

		c.SetUserContext(otelCtx)

		c.Response().Header.Add("X-Trace-ID", span.SpanContext().TraceID().String())

		defer span.End()

		err = c.Next()

		statusCode := c.Response().StatusCode()
		attrs := semconv.HTTPAttributesFromHTTPStatusCode(statusCode)
		spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCodeAndSpanKind(statusCode, trace.SpanKindServer)
		span.SetAttributes(attrs...)
		span.SetStatus(spanStatus, spanMessage)

		return err
	}
}

func concatSpanOptions(sources ...[]trace.SpanStartOption) []trace.SpanStartOption {
	var spanOptions []trace.SpanStartOption
	for _, source := range sources {
		for _, option := range source {
			spanOptions = append(spanOptions, option)
		}
	}

	return spanOptions
}
