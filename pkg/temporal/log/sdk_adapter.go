// Licensed to Alexandre VILAIN under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Alexandre VILAIN licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
