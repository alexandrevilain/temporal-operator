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

package v1beta1

import (
	"testing"
	"time"

	"github.com/alexandrevilain/temporal-operator/pkg/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The default version and the release catalog are bumped together when a new Temporal version is
// supported; target: Latest must not converge somewhere a fresh cluster would not be created at.
func TestDefaultVersionIsLatestKnownRelease(t *testing.T) {
	assert.Equal(t, defaultTemporalVersion, version.LatestKnownRelease().String())
}

func TestDefaultUpgradeSpec(t *testing.T) {
	c := &TemporalCluster{}
	c.Default()
	assert.Nil(t, c.Spec.Upgrade, "spec.upgrade must stay unset so existing objects are not modified")

	c = &TemporalCluster{Spec: TemporalClusterSpec{Upgrade: &UpgradeSpec{}}}
	c.Default()
	assert.Equal(t, ManualUpgradeStrategy, c.Spec.Upgrade.Strategy)
	assert.Empty(t, c.Spec.Upgrade.Target)
	assert.Nil(t, c.Spec.Upgrade.StepTimeout)

	c = &TemporalCluster{Spec: TemporalClusterSpec{Upgrade: &UpgradeSpec{Strategy: AutomaticUpgradeStrategy}}}
	c.Default()
	assert.Equal(t, VersionUpgradeTarget, c.Spec.Upgrade.Target)
	require.NotNil(t, c.Spec.Upgrade.StepTimeout)
	assert.Equal(t, time.Hour, c.Spec.Upgrade.StepTimeout.Duration)
}

func TestUpgradeTargetVersion(t *testing.T) {
	v := version.MustNewVersionFromString

	c := &TemporalCluster{Spec: TemporalClusterSpec{Version: v("1.24.3")}}
	assert.Equal(t, "1.24.3", c.UpgradeTargetVersion().String())

	c.Spec.Upgrade = &UpgradeSpec{Strategy: AutomaticUpgradeStrategy, Target: VersionUpgradeTarget}
	assert.Equal(t, "1.24.3", c.UpgradeTargetVersion().String())

	c.Spec.Upgrade.Target = LatestUpgradeTarget
	assert.Equal(t, version.LatestKnownRelease().String(), c.UpgradeTargetVersion().String())

	// Latest only matters with the automatic strategy.
	c.Spec.Upgrade.Strategy = ManualUpgradeStrategy
	assert.Equal(t, "1.24.3", c.UpgradeTargetVersion().String())
}

func TestObservedVersionFloor(t *testing.T) {
	v := version.MustNewVersionFromString

	c := &TemporalCluster{}
	assert.Nil(t, c.ObservedVersionFloor(), "a cluster never deployed has no floor")

	c.Status.Services = []ServiceStatus{{Name: "frontend", Version: "0.0.0"}}
	assert.Nil(t, c.ObservedVersionFloor(), "services without a version label don't count")

	c.Status.Version = "1.24.3"
	c.Status.Services = []ServiceStatus{{Name: "frontend", Version: "1.24.3"}, {Name: "history", Version: "1.24.3"}}
	assert.Equal(t, "1.24.3", c.ObservedVersionFloor().String())

	// A manual upgrade half way through: the schema has been migrated, deployments not yet.
	c.Status.Persistence = &TemporalPersistenceStatus{
		DefaultStore:    &DatastoreStatus{SchemaVersion: v("1.25.2")},
		VisibilityStore: &DatastoreStatus{SchemaVersion: v("1.24.3")},
	}
	assert.Equal(t, "1.25.2", c.ObservedVersionFloor().String())

	// One deployment already rolled to a newer version.
	c.Status.Services[1].Version = "1.25.3"
	assert.Equal(t, "1.25.3", c.ObservedVersionFloor().String())

	// An automatic upgrade step in flight.
	c.Status.Upgrade = &UpgradeStatus{CurrentStep: v("1.26.3")}
	assert.Equal(t, "1.26.3", c.ObservedVersionFloor().String())
}
