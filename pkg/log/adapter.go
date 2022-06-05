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
	temporallog "go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type logAdapter struct {
	l logr.Logger
}

// NewTemporalLogFromContext creates a new logger adapater for temporal.
func NewTemporalLogFromContext(ctx context.Context) temporallog.Logger {
	return &logAdapter{l: log.FromContext(ctx)}
}

func (a *logAdapter) applyTags(tags ...tag.Tag) logr.Logger {
	log := a.l
	for _, t := range tags {
		log.WithValues(t.Key(), t.Value())
	}
	return log
}

func (a *logAdapter) Debug(msg string, tags ...tag.Tag) {
	a.applyTags(tags...).V(10).Info(msg)
}

func (a *logAdapter) Info(msg string, tags ...tag.Tag) {
	a.applyTags(tags...).V(1).Info(msg)
}

func (a *logAdapter) Warn(msg string, tags ...tag.Tag) {
	a.applyTags(tags...).V(5).Info(msg)
}

func (a *logAdapter) Error(msg string, tags ...tag.Tag) {
	a.applyTags(tags...).V(0).Info(msg)
}

func (a *logAdapter) Fatal(msg string, tags ...tag.Tag) {
	a.applyTags(tags...).V(0).Info(msg)
}
