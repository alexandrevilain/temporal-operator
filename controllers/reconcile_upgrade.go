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

package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	"github.com/alexandrevilain/temporal-operator/pkg/status"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Automatic upgrades (spec.upgrade.strategy: Automatic)
//
// spec.version (or the operator's newest known release, for spec.upgrade.target: Latest) is the
// version the cluster converges to; it is never rewritten. Each reconcile renders the cluster at
// status.upgrade.currentStep instead, through a copy of the object whose spec.version is the step,
// so every resource builder, schema job and status helper works unchanged. Only the copy's status
// is written back.
//
// A step is complete once the cluster has fully rolled out at the step's version: persistence
// jobs done (they run before any workload is touched), status.version equal to the step, and
// every service deployment ready at that version. Only then does the operator pick the next step,
// the newest known non-broken patch of the next minor version, or the target itself once it is at
// most one minor version away.
//
// A step fails when one of its jobs fails or it runs longer than spec.upgrade.stepTimeout. A
// failed step halts the upgrade, and the operator stops reconciling the cluster's workloads so
// nothing is retried behind the user's back. Any spec change clears the failure and retries the
// step, deleting the step's failed jobs so they run again.

const (
	upgradeProgressRequeue = 30 * time.Second
	upgradeAdvanceRequeue  = time.Second
)

var upgradeClock = time.Now

// prepareUpgrade returns the object the reconcile should render: the cluster itself when automatic
// upgrades are off, otherwise a copy rendered at the current upgrade step. halted is true when the
// current step failed and the workloads must not be reconciled.
func (r *TemporalClusterReconciler) prepareUpgrade(ctx context.Context, cluster *v1beta1.TemporalCluster) (working *v1beta1.TemporalCluster, halted bool, err error) {
	if !cluster.Spec.Upgrade.IsAutomatic() {
		// Leaving the automatic strategy is only accepted by the webhook once spec.version is at
		// most one minor version away from what the cluster runs, so dropping the state is safe.
		if cluster.Status.Upgrade != nil {
			cluster.Status.Upgrade = nil
			v1beta1.RemoveTemporalClusterUpgrading(cluster)
		}
		return cluster, false, nil
	}

	if cluster.Status.Upgrade == nil {
		cluster.Status.Upgrade = &v1beta1.UpgradeStatus{}
	}
	u := cluster.Status.Upgrade
	target := cluster.UpgradeTargetVersion()
	floor := cluster.ObservedVersionFloor()

	if u.TargetVersion == nil || !u.TargetVersion.Equal(target.Version) {
		if u.TargetVersion != nil {
			r.upgradeRecorder().Eventf(cluster, corev1.EventTypeNormal, "UpgradeTargetChanged", "Upgrade target changed from %s to %s", u.TargetVersion, target)
		}
		u.TargetVersion = target.DeepCopy()
		u.StartVersion = floor.DeepCopy()
		if u.Phase == v1beta1.UpgradePhaseCompleted {
			u.Phase = ""
		}
	}

	if u.CurrentStep == nil || (floor != nil && u.CurrentStep.LessThan(floor)) {
		// Start from what the cluster has reached, and confirm it is healthy there before moving
		// on. A cluster that was never deployed is installed at the target directly.
		if floor == nil {
			u.CurrentStep = target.DeepCopy()
		} else {
			u.CurrentStep = floor
		}
		u.StepStartTime = &metav1.Time{Time: upgradeClock()}
		u.Phase = v1beta1.UpgradePhaseProgressing
	}

	if u.Phase == v1beta1.UpgradePhaseFailed {
		if u.FailedGeneration == cluster.GetGeneration() {
			return r.upgradeWorkingCopy(cluster), true, nil
		}

		// The spec changed since the failure: retry the step.
		if err := r.deleteFailedJobs(ctx, cluster, u.CurrentStep); err != nil {
			return nil, false, err
		}
		u.Phase = v1beta1.UpgradePhaseProgressing
		u.FailedGeneration = 0
		u.StepStartTime = &metav1.Time{Time: upgradeClock()}
		r.upgradeRecorder().Eventf(cluster, corev1.EventTypeNormal, "UpgradeRetrying", "Retrying upgrade step %s", u.CurrentStep)
	}

	if u.Phase == "" {
		u.Phase = v1beta1.UpgradePhaseProgressing
	}

	u.PendingSteps = pendingSteps(u.CurrentStep, target)

	return r.upgradeWorkingCopy(cluster), false, nil
}

// upgradeRecorder returns the recorder for upgrade events. client-go rate-limits events per source
// and object (a burst of 25, then one every five minutes), and the cluster controller spends that
// budget on an event per child resource it creates or updates, so upgrade events sent from the same
// source are dropped exactly when they matter. A dedicated source gets its own budget.
func (r *TemporalClusterReconciler) upgradeRecorder() record.EventRecorder {
	if r.UpgradeRecorder != nil {
		return r.UpgradeRecorder
	}
	return r.Recorder
}

// upgradeWorkingCopy returns a copy of the cluster rendered at the current upgrade step.
func (r *TemporalClusterReconciler) upgradeWorkingCopy(cluster *v1beta1.TemporalCluster) *v1beta1.TemporalCluster {
	working := cluster.DeepCopy()
	step := cluster.Status.Upgrade.CurrentStep
	working.Spec.Version = step.DeepCopy()

	// An admin tools tag that is an operator default (including one copied into a manifest from
	// the docs, e.g. "1.24.2-tctl-1.18.1-cli-1.0.0") follows the step; a custom tag is honoured.
	if working.Spec.AdminTools != nil && version.IsKnownDefaultAdminToolTag(working.Spec.AdminTools.Version) {
		working.Spec.AdminTools.Version = version.DefaultAdminToolTag(step)
	}

	return working
}

// evaluateUpgrade runs after the working copy has been reconciled. It completes, advances or fails
// the current step and returns how soon the cluster should be reconciled again.
func (r *TemporalClusterReconciler) evaluateUpgrade(ctx context.Context, working *v1beta1.TemporalCluster, persistenceDone bool) (time.Duration, error) {
	u := working.Status.Upgrade
	if u == nil || u.Phase == v1beta1.UpgradePhaseFailed {
		return 0, nil
	}

	step := u.CurrentStep
	target := u.TargetVersion
	paused := working.Spec.Upgrade.Paused

	stepComplete := persistenceDone &&
		working.Status.Version == step.String() &&
		status.IsClusterReady(working)

	if stepComplete {
		if step.GreaterOrEqual(target) {
			if u.Phase != v1beta1.UpgradePhaseCompleted {
				if u.StartVersion != nil && !u.StartVersion.Equal(target.Version) {
					r.upgradeRecorder().Eventf(working, corev1.EventTypeNormal, "UpgradeCompleted", "Upgrade from %s to %s completed", u.StartVersion, step)
				}
			}
			u.Phase = v1beta1.UpgradePhaseCompleted
			u.PendingSteps = nil
			u.Message = fmt.Sprintf("Cluster runs the target version %s", step)
			v1beta1.SetTemporalClusterUpgrading(working, metav1.ConditionFalse, v1beta1.UpgradeCompletedReason, u.Message)
			return 0, nil
		}

		next, err := version.NextUpgradeStep(step, target)
		if err != nil {
			r.failUpgrade(working, fmt.Sprintf("Can't compute the next upgrade step from %s to %s: %v", step, target, err))
			return 0, nil
		}

		if paused {
			u.Phase = v1beta1.UpgradePhasePaused
			u.Message = fmt.Sprintf("Upgrade paused at %s; next step is %s (target %s)", step, next, target)
			v1beta1.SetTemporalClusterUpgrading(working, metav1.ConditionFalse, v1beta1.UpgradePausedReason, u.Message)
			return 0, nil
		}

		u.CurrentStep = next
		u.StepStartTime = &metav1.Time{Time: upgradeClock()}
		u.Phase = v1beta1.UpgradePhaseProgressing
		u.PendingSteps = pendingSteps(next, target)
		u.Message = fmt.Sprintf("Rolling out %s (from %s, target %s)", next, step, target)
		v1beta1.SetTemporalClusterUpgrading(working, metav1.ConditionTrue, v1beta1.UpgradeInProgressReason, u.Message)
		r.upgradeRecorder().Eventf(working, corev1.EventTypeNormal, "UpgradeStepStarted", "Upgrading from %s to %s (target %s)", step, next, target)
		return upgradeAdvanceRequeue, nil
	}

	failedJob, err := r.failedJobForStep(ctx, working, step, u.StepStartTime)
	if err != nil {
		return 0, err
	}
	if failedJob != "" {
		r.failUpgrade(working, fmt.Sprintf("Upgrade step %s failed: job %s failed. Fix the cause, then change the spec (e.g. toggle spec.upgrade.paused) to retry the step", step, failedJob))
		return 0, nil
	}

	timeout := v1beta1.DefaultUpgradeStepTimeout
	if working.Spec.Upgrade.StepTimeout != nil && working.Spec.Upgrade.StepTimeout.Duration > 0 {
		timeout = working.Spec.Upgrade.StepTimeout.Duration
	}
	var elapsed time.Duration
	if u.StepStartTime != nil {
		elapsed = upgradeClock().Sub(u.StepStartTime.Time)
	}
	if elapsed > timeout {
		r.failUpgrade(working, fmt.Sprintf("Upgrade step %s did not complete within %s (%s). Fix the cause, then change the spec (e.g. toggle spec.upgrade.paused) to retry the step", step, timeout, stepProgress(working, persistenceDone)))
		return 0, nil
	}

	u.Phase = v1beta1.UpgradePhaseProgressing
	if step.GreaterOrEqual(target) && u.StartVersion == nil {
		u.Message = fmt.Sprintf("Installing %s (%s)", step, stepProgress(working, persistenceDone))
	} else {
		u.Message = fmt.Sprintf("Rolling out %s, target %s (%s)", step, target, stepProgress(working, persistenceDone))
	}
	if paused {
		u.Message += "; paused, no further step will be started"
	}
	v1beta1.SetTemporalClusterUpgrading(working, metav1.ConditionTrue, v1beta1.UpgradeInProgressReason, u.Message)

	return min(upgradeProgressRequeue, timeout-elapsed+time.Second), nil
}

func (r *TemporalClusterReconciler) failUpgrade(working *v1beta1.TemporalCluster, message string) {
	u := working.Status.Upgrade
	u.Phase = v1beta1.UpgradePhaseFailed
	u.FailedGeneration = working.GetGeneration()
	u.Message = message
	v1beta1.SetTemporalClusterUpgrading(working, metav1.ConditionFalse, v1beta1.UpgradeFailedReason, message)
	r.upgradeRecorder().Event(working, corev1.EventTypeWarning, "UpgradeFailed", message)
}

// stepProgress describes what a step is waiting for.
func stepProgress(working *v1beta1.TemporalCluster, persistenceDone bool) string {
	if !persistenceDone {
		return "waiting for persistence schema jobs"
	}
	notReady := []string{}
	for _, svc := range working.Status.Services {
		if !svc.Ready || svc.Version != working.Spec.Version.String() {
			notReady = append(notReady, svc.Name)
		}
	}
	if len(notReady) == 0 {
		return "waiting for services to report ready"
	}
	return "waiting for services: " + strings.Join(notReady, ", ")
}

func pendingSteps(current, target *version.Version) []string {
	path, err := version.UpgradePath(current, target)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(path))
	for _, v := range path {
		out = append(out, v.String())
	}
	return out
}

// ownedJobsForVersion lists the cluster's jobs rendered for the given version.
func (r *TemporalClusterReconciler) ownedJobsForVersion(ctx context.Context, cluster *v1beta1.TemporalCluster, v *version.Version) ([]batchv1.Job, error) {
	selector := client.MatchingLabels{}
	for k, val := range cluster.SelectorLabels() {
		selector[k] = val
	}
	selector["app.kubernetes.io/version"] = v.String()

	jobs := &batchv1.JobList{}
	if err := r.List(ctx, jobs, client.InNamespace(cluster.Namespace), selector); err != nil {
		return nil, fmt.Errorf("can't list jobs: %w", err)
	}
	out := []batchv1.Job{}
	for _, job := range jobs.Items {
		// Schema jobs carry a plain (non-controller) owner reference to the cluster.
		for _, ref := range job.OwnerReferences {
			if ref.UID == cluster.UID {
				out = append(out, job)
				break
			}
		}
	}
	return out, nil
}

// jobFailureTime returns when the job failed, or nil if it has not.
func jobFailureTime(job *batchv1.Job) *metav1.Time {
	for _, c := range job.Status.Conditions {
		if c.Type == batchv1.JobFailed && c.Status == corev1.ConditionTrue {
			return &c.LastTransitionTime
		}
	}
	return nil
}

func isJobFailed(job *batchv1.Job) bool {
	return jobFailureTime(job) != nil
}

// failedJobForStep returns the name of a job of the step that failed after the step started (or
// was last retried), if any. Older failures are left to the step timeout: they are either jobs
// being deleted for a retry, which the cache may still serve, or leftovers from before the
// automatic upgrade took over.
func (r *TemporalClusterReconciler) failedJobForStep(ctx context.Context, cluster *v1beta1.TemporalCluster, step *version.Version, since *metav1.Time) (string, error) {
	jobs, err := r.ownedJobsForVersion(ctx, cluster, step)
	if err != nil {
		return "", err
	}
	for i := range jobs {
		if !jobs[i].DeletionTimestamp.IsZero() {
			continue
		}
		failedAt := jobFailureTime(&jobs[i])
		if failedAt == nil {
			continue
		}
		if since != nil && failedAt.Before(since) {
			continue
		}
		return jobs[i].Name, nil
	}
	return "", nil
}

// deleteFailedJobs deletes the step's failed jobs so that the persistence reconciler, which only
// creates missing jobs, runs them again.
func (r *TemporalClusterReconciler) deleteFailedJobs(ctx context.Context, cluster *v1beta1.TemporalCluster, step *version.Version) error {
	jobs, err := r.ownedJobsForVersion(ctx, cluster, step)
	if err != nil {
		return err
	}
	for i := range jobs {
		if !isJobFailed(&jobs[i]) {
			continue
		}
		log.FromContext(ctx).Info("Deleting failed upgrade job to retry it", "job", jobs[i].Name)
		err := r.Delete(ctx, &jobs[i], client.PropagationPolicy(metav1.DeletePropagationBackground))
		if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("can't delete failed job %s: %w", jobs[i].Name, err)
		}
	}
	return nil
}
