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

package e2e

import (
	"context"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
)

type (
	temporalClusterContextKey       string
	temporalClusterClientContextKey string
	temporalNamespaceContextKey     string

	namespaceContextKey string
)

var (
	temporalClusterKey       temporalClusterContextKey       = "temporalCluster"
	temporalClusterClientKey temporalClusterClientContextKey = "temporalClusterClient"
	temporalNamespaceKey     temporalNamespaceContextKey     = "temporalNamespace"
	temporalScheduleKey      temporalNamespaceContextKey     = "temporalSchedule"

	namespaceKey namespaceContextKey = "namespace"
)

func GetNamespaceForFeature(ctx context.Context) string {
	return ctx.Value(namespaceKey).(string)
}

func SetNamespaceForFeature(ctx context.Context, namespace string) context.Context {
	return context.WithValue(ctx, namespaceKey, namespace)
}

func GetTemporalClusterForFeature(ctx context.Context) *v1beta1.TemporalCluster {
	return ctx.Value(temporalClusterKey).(*v1beta1.TemporalCluster)
}

func SetTemporalClusterForFeature(ctx context.Context, cluster *v1beta1.TemporalCluster) context.Context {
	return context.WithValue(ctx, temporalClusterKey, cluster)
}

func GetTemporalClusterClientForFeature(ctx context.Context) *v1beta1.TemporalClusterClient {
	return ctx.Value(temporalClusterClientKey).(*v1beta1.TemporalClusterClient)
}

func SetTemporalClusterClientForFeature(ctx context.Context, clusterClient *v1beta1.TemporalClusterClient) context.Context {
	return context.WithValue(ctx, temporalClusterClientKey, clusterClient)
}

func GetTemporalNamespaceForFeature(ctx context.Context) *v1beta1.TemporalNamespace {
	return ctx.Value(temporalNamespaceKey).(*v1beta1.TemporalNamespace)
}

func SetTemporalNamespaceForFeature(ctx context.Context, namespace *v1beta1.TemporalNamespace) context.Context {
	return context.WithValue(ctx, temporalNamespaceKey, namespace)
}

func GetTemporalScheduleForFeature(ctx context.Context) *v1beta1.TemporalSchedule {
	return ctx.Value(temporalScheduleKey).(*v1beta1.TemporalSchedule)
}

func SetTemporalScheduleForFeature(ctx context.Context, Schedule *v1beta1.TemporalSchedule) context.Context {
	return context.WithValue(ctx, temporalScheduleKey, Schedule)
}
