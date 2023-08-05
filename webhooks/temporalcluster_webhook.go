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

package webhooks

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/discovery"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	enumspb "go.temporal.io/api/enums/v1"
	enumsspb "go.temporal.io/server/api/enums/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/pointer"
	"k8s.io/utils/strings/slices"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// TemporalClusterWebhook provides endpoints to validate
// and set default fields values for TemporalCluster objects.
type TemporalClusterWebhook struct {
	AvailableAPIs *discovery.AvailableAPIs
}

func (w *TemporalClusterWebhook) getClusterFromRequest(obj runtime.Object) (*v1beta1.TemporalCluster, error) {
	cluster, ok := obj.(*v1beta1.TemporalCluster)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected an TemporalCluster but got a %T", obj))
	}
	return cluster, nil
}

func (w *TemporalClusterWebhook) aggregateClusterErrors(cluster *v1beta1.TemporalCluster, errs field.ErrorList) error {
	if len(errs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		cluster.GroupVersionKind().GroupKind(),
		cluster.GetName(),
		errs,
	)
}

// Default ensures empty fields have their default value.
func (w *TemporalClusterWebhook) Default(ctx context.Context, obj runtime.Object) error {
	cluster, err := w.getClusterFromRequest(obj)
	if err != nil {
		return err
	}

	if cluster.Spec.Metrics.IsEnabled() {
		if cluster.Spec.Metrics.Prometheus != nil {
			// If the user has set the deprecated ListenAddress field and not the new ListenPort,
			// parse the listenAddress and set the listenPort.
			if cluster.Spec.Metrics.Prometheus.ListenAddress != "" && cluster.Spec.Metrics.Prometheus.ListenPort == nil {
				_, port, err := net.SplitHostPort(cluster.Spec.Metrics.Prometheus.ListenAddress)
				if err != nil {
					return fmt.Errorf("can't parse prometheus spec.metrics.prometheus.listenAddress: %w", err)
				}
				portInt, err := strconv.ParseInt(port, 10, 32)
				if err != nil {
					return fmt.Errorf("can't parse prometheus spec.metrics.prometheus.listenAddress port: %w", err)
				}
				cluster.Spec.Metrics.Prometheus.ListenAddress = "" // Empty the listen address
				cluster.Spec.Metrics.Prometheus.ListenPort = pointer.Int32(int32(portInt))
			}
		}
	}

	// Finish by setting default values
	cluster.Default()

	return nil
}

func (w *TemporalClusterWebhook) validateCluster(cluster *v1beta1.TemporalCluster) (admission.Warnings, field.ErrorList) {
	var warns admission.Warnings
	var errs field.ErrorList

	// If mTLS is enabled using cert-manager, but cert-manager support is disabled on the controller
	// it can't process the request, return the error.
	if cluster.MTLSWithCertManagerEnabled() && !w.AvailableAPIs.CertManager {
		errs = append(errs,
			field.Invalid(
				field.NewPath("spec", "mTLS", "provider"),
				cluster.Spec.MTLS.Provider,
				"Can't use cert-manager as mTLS provider as it's not available in the cluster",
			),
		)
	}

	// Validate that the cluster version is a supported one.
	err := cluster.Spec.Version.Validate()
	if err != nil {
		errs = append(errs,
			field.Forbidden(
				field.NewPath("spec", "version"),
				fmt.Sprintf("Unsupported temporal version (supported: %s)", version.SupportedVersionsRange.String()),
			),
		)
	}

	// Ensure ElasticSearch v6 is not used with cluster >= 1.18.0
	if cluster.Spec.Version.GreaterOrEqual(version.V1_18_0) &&
		cluster.Spec.Persistence.AdvancedVisibilityStore != nil &&
		cluster.Spec.Persistence.AdvancedVisibilityStore.Elasticsearch != nil &&
		cluster.Spec.Persistence.AdvancedVisibilityStore.Elasticsearch.Version == "v6" {
		errs = append(errs,
			field.Forbidden(
				field.NewPath("spec", "persistence", "advancedVisibilityStore", "elasticsearch", "version"),
				"temporal cluster version >= 1.18.0 doesn't support ElasticSearch v6",
			),
		)
	}

	// Ensure dynamicconfig is valid.
	if cluster.Spec.DynamicConfig != nil {
		for key, constrainedValues := range cluster.Spec.DynamicConfig.Values {
			for i, constrainedValue := range constrainedValues {
				c := constrainedValue.Constraints
				if c.TaskQueueType != "" {
					if _, ok := enumspb.TaskQueueType_value[c.TaskQueueType]; !ok {
						supportedValues := []string{}
						for k, v := range enumspb.TaskQueueType_name {
							if k == 0 {
								continue
							}
							supportedValues = append(supportedValues, v)
						}
						errs = append(errs,
							field.NotSupported(
								field.NewPath("spec", "dynamicConfig", "values", key, fmt.Sprintf("[%d]", i), "constraints", "taskQueueType"),
								c.TaskQueueType,
								supportedValues,
							),
						)
					}
				}

				if c.TaskType != "" {
					if _, ok := enumsspb.TaskType_value[c.TaskType]; !ok {
						supportedValues := []string{}
						for k, v := range enumsspb.TaskType_name {
							if k == 0 {
								continue
							}
							supportedValues = append(supportedValues, v)
						}
						errs = append(errs,
							field.NotSupported(
								field.NewPath("spec", "dynamicConfig", "values", key, fmt.Sprintf("[%d]", i), "constraints", ".taskType"),
								c.TaskType,
								supportedValues,
							),
						)
					}
				}
			}
		}
	}

	// Check that the user-specified version is not marked as broken.
	for _, version := range version.ForbiddenBrokenReleases {
		if cluster.Spec.Version.Equal(version.Version) {
			errs = append(errs,
				field.Forbidden(
					field.NewPath("spec", "version"),
					fmt.Sprintf("version %s is marked as broken by the operator, please upgrade to %s (if allowed)", cluster.Spec.Version.String(), cluster.Spec.Version.IncPatch().String()),
				),
			)
		}
	}

	// Check new features introduced in cluster version >= 1.20 are not enabled for older version.
	if !cluster.Spec.Version.GreaterOrEqual(version.V1_20_0) {
		// Ensure Internal Frontend is only enabled for cluster version >= 1.20
		if cluster.Spec.Services != nil && cluster.Spec.Services.InternalFrontend.IsEnabled() {
			errs = append(errs,
				field.Forbidden(
					field.NewPath("spec", "services", "internalFrontend", "enabled"),
					"temporal cluster version < 1.20.0 doesn't support internal frontend",
				),
			)
		}

		// Ensure mysql8 and postgres12 plugins are only used for cluster version >= 1.20
		newStores := []string{string(v1beta1.PostgresSQL12Datastore), string(v1beta1.MySQL8Datastore)}
		for name, store := range cluster.Spec.Persistence.GetDatastoresMap() {
			if store != nil && store.SQL != nil && slices.Contains(newStores, store.SQL.PluginName) {
				errs = append(errs,
					field.Forbidden(
						field.NewPath("spec", "persistence", name, "sql", "pluginName"),
						fmt.Sprintf("temporal cluster version < 1.20.0 doesn't support %s plugin name", store.SQL.PluginName),
					),
				)
			}
		}
	}

	// Check new features introduced in cluster version >= 1.21 are not enabled for older version.
	if !cluster.Spec.Version.GreaterOrEqual(version.V1_21_0) {
		if cluster.Spec.Persistence.SecondaryVisibilityStore != nil {
			errs = append(errs,
				field.Forbidden(
					field.NewPath("spec", "persistence", "secondaryVisibilityStore"),
					"temporal cluster version < 1.21.0 doesn't support secondary visibility store",
				),
			)
		}
	}

	// Check for visibility store depreciations introduced in >= 1.21, that will be removed in >=1.23
	if cluster.Spec.Version.GreaterOrEqual(version.V1_21_0) {
		if cluster.Spec.Persistence.AdvancedVisibilityStore != nil {
			warns = append(warns,
				"Starting from temporal >= 1.21 standard visibility becomes advanced visibility. Advanced visibility configuration is now moved to standard visibility. Please only use visibility datastore configuration. Avanced visibility store usage will be forbidden by the operator for clusters >= 1.23.",
			)
		}

		if cluster.Spec.Persistence.VisibilityStore != nil && cluster.Spec.Persistence.VisibilityStore.Cassandra != nil {
			warns = append(warns,
				"Support for Cassandra as a Visibility database is deprecated beginning with Temporal Server v1.21.",
			)
		}
	}

	return warns, errs
}

// ValidateCreate ensures the user is creating a consistent temporal cluster.
func (w *TemporalClusterWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	cluster, err := w.getClusterFromRequest(obj)
	if err != nil {
		return nil, err
	}

	warns, errs := w.validateCluster(cluster)

	return warns, w.aggregateClusterErrors(cluster, errs)
}

// ValidateUpdate validates TemporalCluster updates.
// It mainly check for sequential version upgrades.
func (w *TemporalClusterWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	oldCluster, err := w.getClusterFromRequest(oldObj)
	if err != nil {
		return nil, err
	}

	newCluster, err := w.getClusterFromRequest(newObj)
	if err != nil {
		return nil, err
	}

	warns, errs := w.validateCluster(newCluster)

	// Ensure user is doing a sequential version upgrade.
	// See: https://docs.temporal.io/cluster-deployment-guide#upgrade-server
	constraint, err := oldCluster.Spec.Version.UpgradeConstraint()
	if err != nil {
		return nil, fmt.Errorf("can't compute version upgrade constraint: %w", err)
	}

	allowed := constraint.Check(newCluster.Spec.Version.Version)
	if !allowed {
		errs = append(errs,
			field.Forbidden(
				field.NewPath("spec", "version"),
				"Unauthorized version upgrade. Only sequential version upgrades are allowed (from v1.n.x to v1.n+1.x)",
			),
		)
	}

	// Ensure user can't update the spec.numHistoryShards.
	// In a temporal cluster, the number of shards is set once and forever.
	if newCluster.Spec.NumHistoryShards != oldCluster.Spec.NumHistoryShards {
		errs = append(errs,
			field.Forbidden(
				field.NewPath("spec", "numHistoryShards"),
				"Number of history shards is immutable",
			),
		)
	}

	return warns, w.aggregateClusterErrors(newCluster, errs)
}

// ValidateDelete does nothing.
func (w *TemporalClusterWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	// No delete validation needed.
	return nil, nil
}

func (w *TemporalClusterWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1beta1.TemporalCluster{}).
		WithDefaulter(w).
		WithValidator(w).
		Complete()
}
