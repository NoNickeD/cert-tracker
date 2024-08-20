package tracing

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

func InitTracer() (*sdktrace.TracerProvider, error) {
	exporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("cert-tracker"),
		)),
	)
	otel.SetTracerProvider(tp)

	return tp, nil
}

func LogError(ctx context.Context, message string, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	span.SetStatus(codes.Error, message)
	logULFF("ERROR", "cert-tracker", message, map[string]interface{}{"error": err.Error()})
}

func LogInfo(ctx context.Context, message string, attributes map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	for k, v := range attributes {
		span.SetAttributes(attribute.String(k, fmt.Sprintf("%v", v)))
	}
	span.AddEvent(message)
	logULFF("INFO", "cert-tracker", message, attributes)
}

func logULFF(level, component, message string, contextInfo map[string]interface{}) {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	contextStr, _ := json.Marshal(contextInfo)
	log.Printf("%s [%s] [%s] %s Context: %s\n", timestamp, level, component, message, contextStr)
}
