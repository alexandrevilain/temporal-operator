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

type serverLogAdapter struct {
	l logr.Logger
}

// NewTemporalServerLogFromContext creates a new logger adapater for temporal.
func NewTemporalServerLogFromContext(ctx context.Context) temporallog.Logger {
	return &serverLogAdapter{l: log.FromContext(ctx)}
}

func (a *serverLogAdapter) applyTags(tags ...tag.Tag) logr.Logger {
	log := a.l
	for _, t := range tags {
		log.WithValues(t.Key(), t.Value())
	}
	return log
}

func (a *serverLogAdapter) Debug(msg string, tags ...tag.Tag) {
	a.applyTags(tags...).V(10).Info(msg)
}

func (a *serverLogAdapter) Info(msg string, tags ...tag.Tag) {
	a.applyTags(tags...).V(1).Info(msg)
}

func (a *serverLogAdapter) Warn(msg string, tags ...tag.Tag) {
	a.applyTags(tags...).V(2).Info(msg)
}

func (a *serverLogAdapter) Error(msg string, tags ...tag.Tag) {
	a.applyTags(tags...).V(0).Info(msg)
}

func (a *serverLogAdapter) Fatal(msg string, tags ...tag.Tag) {
	a.applyTags(tags...).V(0).Info(msg)
}
