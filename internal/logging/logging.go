package logging

import (
	"context"

	liblogger "github.com/Kaese72/huemie-lib/logging"
	"github.com/Kaese72/huemie-lib/telemetry"
)

func traceFields(ctx context.Context) map[string]interface{} {
	return telemetry.LogFields(ctx)
}

func Info(msg string, ctx context.Context, data ...map[string]interface{}) {
	liblogger.Info(msg, append(data, traceFields(ctx))...)
}

func Error(msg string, ctx context.Context, data ...map[string]interface{}) {
	liblogger.Error(msg, append(data, traceFields(ctx))...)
}

func ErrorErr(err error, ctx context.Context, data ...map[string]interface{}) {
	liblogger.ErrError(err, append(data, traceFields(ctx))...)
}

func Fatal(msg string, ctx context.Context, data ...map[string]interface{}) {
	liblogger.Fatal(msg, append(data, traceFields(ctx))...)
}
