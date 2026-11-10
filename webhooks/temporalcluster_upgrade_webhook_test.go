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

package webhooks_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/internal/discovery"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"github.com/alexandrevilain/temporal-operator/webhooks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// upgradeCluster returns a defaulted cluster at specVersion that the operator has rolled out at
// runningVersion ("" for a cluster never deployed).
func upgradeCluster(specVersion, runningVersion string, upgrade *v1beta1.UpgradeSpec) *v1beta1.TemporalCluster {
	c := &v1beta1.TemporalCluster{
		TypeMeta:   v1beta1.TemporalClusterTypeMeta,
		ObjectMeta: metav1.ObjectMeta{Name: "fake"},
		Spec: v1beta1.TemporalClusterSpec{
			Version:          version.MustNewVersionFromString(specVersion),
			NumHistoryShards: 512,
			Upgrade:          upgrade,
		},
	}
	if runningVersion != "" {
		c.Status.Version = runningVersion
		c.Status.Services = []v1beta1.ServiceStatus{{Name: "frontend", Version: runningVersion, Ready: true}}
	}
	c.Default()
	return c
}

func automatic() *v1beta1.UpgradeSpec {
	return &v1beta1.UpgradeSpec{Strategy: v1beta1.AutomaticUpgradeStrategy}
}

func latest() *v1beta1.UpgradeSpec {
	return &v1beta1.UpgradeSpec{Strategy: v1beta1.AutomaticUpgradeStrategy, Target: v1beta1.LatestUpgradeTarget}
}

// with1274InFlight marks an automatic upgrade step to 1.27.4 as in flight.
func with1274InFlight(c *v1beta1.TemporalCluster) *v1beta1.TemporalCluster {
	c.Status.Upgrade = &v1beta1.UpgradeStatus{
		Phase:       v1beta1.UpgradePhaseProgressing,
		CurrentStep: version.MustNewVersionFromString("1.27.4"),
	}
	return c
}

func TestValidateUpdate_AutomaticUpgrade(t *testing.T) {
	tests := map[string]struct {
		old, new    *v1beta1.TemporalCluster
		expectedErr string
	}{
		"enabling automatic upgrades lets spec.version skip minor versions": {
			old: upgradeCluster("1.24.3", "1.24.3", nil),
			new: upgradeCluster("1.32.2", "1.24.3", automatic()),
		},
		"the same jump without the automatic strategy is still refused": {
			old:         upgradeCluster("1.24.3", "1.24.3", nil),
			new:         upgradeCluster("1.32.2", "1.24.3", nil),
			expectedErr: "Only sequential version upgrades are allowed",
		},
		"an explicit manual strategy keeps the sequential rule": {
			old:         upgradeCluster("1.24.3", "1.24.3", nil),
			new:         upgradeCluster("1.26.3", "1.24.3", &v1beta1.UpgradeSpec{Strategy: v1beta1.ManualUpgradeStrategy}),
			expectedErr: "Only sequential version upgrades are allowed",
		},
		"enabling the Latest target": {
			old: upgradeCluster("1.24.3", "1.24.3", nil),
			new: upgradeCluster("1.24.3", "1.24.3", latest()),
		},
		"raising the target mid-upgrade": {
			old: with1274InFlight(upgradeCluster("1.29.7", "1.26.3", automatic())),
			new: upgradeCluster("1.32.2", "1.26.3", automatic()),
		},
		"lowering the target to the step in flight": {
			old: with1274InFlight(upgradeCluster("1.32.2", "1.26.3", automatic())),
			new: upgradeCluster("1.27.4", "1.26.3", automatic()),
		},
		"lowering the target below the step in flight is a downgrade": {
			old:         with1274InFlight(upgradeCluster("1.32.2", "1.26.3", automatic())),
			new:         upgradeCluster("1.26.3", "1.26.3", automatic()),
			expectedErr: "the upgrade target 1.26.3 is lower than 1.27.4",
		},
		"automatic upgrades can't downgrade the running version": {
			old:         upgradeCluster("1.28.4", "1.28.4", automatic()),
			new:         upgradeCluster("1.27.4", "1.28.4", automatic()),
			expectedErr: "downgrades are not supported",
		},
		"automatic upgrades from a cluster never deployed use the old spec.version as floor": {
			old:         upgradeCluster("1.28.4", "", nil),
			new:         upgradeCluster("1.27.4", "", automatic()),
			expectedErr: "downgrades are not supported",
		},
		"leaving the automatic strategy with spec.version far ahead is refused": {
			old:         with1274InFlight(upgradeCluster("1.32.2", "1.26.3", automatic())),
			new:         upgradeCluster("1.32.2", "1.26.3", nil),
			expectedErr: "Can't leave the Automatic upgrade strategy with spec.version 1.32.2: the cluster has reached 1.27.4",
		},
		"leaving the automatic strategy with a sequential spec.version": {
			old: with1274InFlight(upgradeCluster("1.32.2", "1.26.3", automatic())),
			new: upgradeCluster("1.28.4", "1.26.3", &v1beta1.UpgradeSpec{Strategy: v1beta1.ManualUpgradeStrategy}),
		},
		"leaving the Latest target with a stale spec.version is refused": {
			old:         upgradeCluster("1.24.3", "1.32.2", latest()),
			new:         upgradeCluster("1.24.3", "1.32.2", nil),
			expectedErr: "the cluster has reached 1.32.2",
		},
		"leaving the Latest target for the version it reached": {
			old: upgradeCluster("1.24.3", "1.32.2", latest()),
			new: upgradeCluster("1.32.2", "1.32.2", automatic()),
		},
		"leaving the Latest target for a stale spec.version under the automatic strategy is a downgrade": {
			old:         upgradeCluster("1.24.3", "1.32.2", latest()),
			new:         upgradeCluster("1.24.3", "1.32.2", automatic()),
			expectedErr: "downgrades are not supported",
		},
	}

	for name, test := range tests {
		t.Run(name, func(tt *testing.T) {
			wh := &webhooks.TemporalClusterWebhook{AvailableAPIs: &discovery.AvailableAPIs{}}
			_, err := wh.ValidateUpdate(context.Background(), test.old, test.new)
			if test.expectedErr != "" {
				require.Error(tt, err)
				assert.Contains(tt, err.Error(), test.expectedErr)
				return
			}
			assert.NoError(tt, err)
		})
	}
}

func TestValidate_UpgradeSpec(t *testing.T) {
	wh := &webhooks.TemporalClusterWebhook{AvailableAPIs: &discovery.AvailableAPIs{}}

	t.Run("Latest target requires the automatic strategy", func(tt *testing.T) {
		c := upgradeCluster("1.32.2", "", &v1beta1.UpgradeSpec{Strategy: v1beta1.ManualUpgradeStrategy, Target: v1beta1.LatestUpgradeTarget})
		_, err := wh.ValidateCreate(context.Background(), c)
		require.Error(tt, err)
		assert.Contains(tt, err.Error(), `spec.upgrade.target: Forbidden: target "Latest" requires strategy "Automatic"`)
	})

	t.Run("stepTimeout has a lower bound", func(tt *testing.T) {
		c := upgradeCluster("1.32.2", "", automatic())
		c.Spec.Upgrade.StepTimeout = &metav1.Duration{Duration: 10 * time.Second}
		_, err := wh.ValidateCreate(context.Background(), c)
		require.Error(tt, err)
		assert.Contains(tt, err.Error(), "spec.upgrade.stepTimeout")
	})

	t.Run("paused without the automatic strategy warns", func(tt *testing.T) {
		c := upgradeCluster("1.32.2", "", &v1beta1.UpgradeSpec{Paused: true})
		warns, err := wh.ValidateCreate(context.Background(), c)
		require.NoError(tt, err)
		assert.Contains(tt, strings.Join(warns, "\n"), "spec.upgrade.paused has no effect")
	})

	t.Run("the Latest target warns that operator upgrades may upgrade Temporal", func(tt *testing.T) {
		c := upgradeCluster("1.32.2", "", latest())
		warns, err := wh.ValidateCreate(context.Background(), c)
		require.NoError(tt, err)
		assert.Contains(tt, strings.Join(warns, "\n"), "upgrading the operator may start a new multi-step Temporal upgrade")
	})

	t.Run("a custom admin tools tag warns that it stays frozen", func(tt *testing.T) {
		c := upgradeCluster("1.32.2", "", automatic())
		c.Spec.AdminTools.Version = "my-mirror-build"
		warns, err := wh.ValidateCreate(context.Background(), c)
		require.NoError(tt, err)
		assert.Contains(tt, strings.Join(warns, "\n"), `spec.admintools.version "my-mirror-build" is not an operator default`)
	})

	t.Run("a default admin tools tag copied from the docs does not warn", func(tt *testing.T) {
		c := upgradeCluster("1.32.2", "", automatic())
		c.Spec.AdminTools.Version = "1.24.2-tctl-1.18.1-cli-1.0.0"
		warns, err := wh.ValidateCreate(context.Background(), c)
		require.NoError(tt, err)
		assert.NotContains(tt, strings.Join(warns, "\n"), "spec.admintools.version")
	})
}

// Rules that depend on the server version must hold at every intermediate step, not only at the
// target: sql.passwordCommand requires 1.31, so an automatic upgrade from 1.29 to 1.32 would render
// it on 1.30 first.
func TestValidateUpdate_AutomaticUpgradeValidatesIntermediateSteps(t *testing.T) {
	wh := &webhooks.TemporalClusterWebhook{AvailableAPIs: &discovery.AvailableAPIs{}}

	withPasswordCommand := func(c *v1beta1.TemporalCluster) *v1beta1.TemporalCluster {
		c.Spec.Persistence = v1beta1.TemporalPersistenceSpec{
			DefaultStore: &v1beta1.DatastoreSpec{
				Name: "default",
				SQL: &v1beta1.SQLSpec{
					User:            "temporal",
					PluginName:      "postgres12",
					DatabaseName:    "temporal",
					ConnectAddr:     "postgres:5432",
					PasswordCommand: &v1beta1.SQLPasswordCommandSpec{Command: "/usr/local/bin/token"},
				},
			},
			VisibilityStore: &v1beta1.DatastoreSpec{
				Name:              "visibility",
				SQL:               &v1beta1.SQLSpec{User: "temporal", PluginName: "postgres12", DatabaseName: "temporal_visibility", ConnectAddr: "postgres:5432"},
				PasswordSecretRef: &v1beta1.SecretKeyReference{Name: "pg", Key: "PASSWORD"},
			},
		}
		return c
	}

	old := upgradeCluster("1.29.7", "1.29.7", nil)
	_, err := wh.ValidateUpdate(context.Background(), old, withPasswordCommand(upgradeCluster("1.32.2", "1.29.7", automatic())))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at upgrade step 1.30.9: sql.passwordCommand requires Temporal >= 1.31.0.")
	assert.NotContains(t, err.Error(), "at upgrade step 1.31.5", "steps where the rule holds must not be reported")

	// Once the cluster runs 1.31, the same spec is accepted.
	old = upgradeCluster("1.31.5", "1.31.5", nil)
	_, err = wh.ValidateUpdate(context.Background(), old, withPasswordCommand(upgradeCluster("1.32.2", "1.31.5", automatic())))
	assert.NoError(t, err)
}
