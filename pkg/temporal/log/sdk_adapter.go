package log

import (
	"context"

	"github.com/go-logr/logr"
	sdklog "go.temporal.io/sdk/log"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Logger interface {
	Debug(msg string, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Warn(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})
}

type sdkLogAdapter struct {
	l logr.Logger
}

// NewTemporalSDKLogFromContext creates a new logger adapater for temporal.
func NewTemporalSDKLogFromContext(ctx context.Context) sdklog.Logger {
	return &sdkLogAdapter{l: log.FromContext(ctx)}
}

func (a *sdkLogAdapter) Debug(msg string, keyvals ...interface{}) {
	a.l.V(10).Info(msg, keyvals...)
}

func (a *sdkLogAdapter) Info(msg string, keyvals ...interface{}) {
	a.l.V(0).Info(msg, keyvals...)
}

func (a *sdkLogAdapter) Warn(msg string, keyvals ...interface{}) {
	a.l.V(5).Info(msg, keyvals...)
}

func (a *sdkLogAdapter) Error(msg string, keyvals ...interface{}) {
	a.l.V(0).Info(msg)
}
