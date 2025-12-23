package logger

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
	"os"
)

func Init() {
	// Set time format to Unix timestamp for better JSON compatibility
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Create a multi-writer for both console (colored) and JSON output
	consoleWriter := zerolog.NewConsoleWriter()

	// Use JSON output when not in development mode
	if os.Getenv("ENV") != "development" {
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	} else {
		log.Logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
	}
}

// Helper functions for structured logging
func Info() *zerolog.Event {
	return addTraceContext(log.Info())
}

func Error() *zerolog.Event {
	return addTraceContext(log.Error())
}

func Debug() *zerolog.Event {
	return addTraceContext(log.Debug())
}

func Warn() *zerolog.Event {
	return addTraceContext(log.Warn())
}

func Fatal() *zerolog.Event {
	return addTraceContext(log.Fatal())
}

// addTraceContext extracts trace ID and span ID from context and adds them to the log event
func addTraceContext(event *zerolog.Event) *zerolog.Event {
	spanCtx := trace.SpanFromContext(context.Background()).SpanContext()
	if spanCtx.IsValid() {
		event.Str("trace_id", spanCtx.TraceID().String())
		event.Str("span_id", spanCtx.SpanID().String())
	}
	return event
}

// CtxInfo creates an info event with trace context from a specific context
func CtxInfo(ctx context.Context) *zerolog.Event {
	return addTraceContextFromCtx(ctx, log.Info())
}

// CtxError creates an error event with trace context from a specific context
func CtxError(ctx context.Context) *zerolog.Event {
	return addTraceContextFromCtx(ctx, log.Error())
}

// CtxDebug creates a debug event with trace context from a specific context
func CtxDebug(ctx context.Context) *zerolog.Event {
	return addTraceContextFromCtx(ctx, log.Debug())
}

// CtxWarn creates a warn event with trace context from a specific context
func CtxWarn(ctx context.Context) *zerolog.Event {
	return addTraceContextFromCtx(ctx, log.Warn())
}

// CtxFatal creates a fatal event with trace context from a specific context
func CtxFatal(ctx context.Context) *zerolog.Event {
	return addTraceContextFromCtx(ctx, log.Fatal())
}

// addTraceContextFromCtx extracts trace context from provided context
func addTraceContextFromCtx(ctx context.Context, event *zerolog.Event) *zerolog.Event {
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	if spanCtx.IsValid() {
		event.Str("trace_id", spanCtx.TraceID().String())
		event.Str("span_id", spanCtx.SpanID().String())
	}
	return event
}
