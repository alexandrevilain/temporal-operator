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

	"github.com/alexandrevilain/controller-tools/pkg/discovery"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/alexandrevilain/temporal-operator/api/v1beta1"
	internaldiscovery "github.com/alexandrevilain/temporal-operator/internal/discovery"
	"github.com/alexandrevilain/temporal-operator/pkg/version"
)

// These specs drive the real TemporalCluster reconciler against envtest. envtest runs no
// kube-controller-manager, so the tests play its part: they complete (or fail) jobs and mark
// deployments as rolled out, which is exactly the signal the upgrade state machine waits for.
var _ = Describe("TemporalCluster automatic upgrades", func() {
	var (
		ctx        context.Context
		reconciler *TemporalClusterReconciler
		recorder   *record.FakeRecorder
		namespace  string
		key        types.NamespacedName
	)

	BeforeEach(func() {
		ctx = context.Background()
		upgradeClock = time.Now

		dm, err := discovery.NewManager(cfg, scheme.Scheme)
		Expect(err).NotTo(HaveOccurred())
		recorder = record.NewFakeRecorder(10000)
		reconciler = &TemporalClusterReconciler{
			// Per-resource events go to their own recorder: the specs assert that upgrade events
			// are recorded through UpgradeRecorder, so they aren't rate-limited away behind them.
			Base:            New(cacheLikeClient{Client: k8sClient}, scheme.Scheme, record.NewFakeRecorder(100000), dm),
			AvailableAPIs:   &internaldiscovery.AvailableAPIs{},
			UpgradeRecorder: recorder,
		}

		namespace = "upgrade-" + rand.String(6)
		Expect(k8sClient.Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}})).To(Succeed())
		key = types.NamespacedName{Namespace: namespace, Name: "cluster"}
	})

	AfterEach(func() {
		upgradeClock = time.Now
	})

	reconcileOnce := func() ctrl.Result {
		GinkgoHelper()
		result, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: key})
		Expect(err).NotTo(HaveOccurred())
		return result
	}

	getCluster := func() *v1beta1.TemporalCluster {
		GinkgoHelper()
		c := &v1beta1.TemporalCluster{}
		Expect(k8sClient.Get(ctx, key, c)).To(Succeed())
		return c
	}

	// createRunningCluster creates a cluster whose persistence the operator has migrated to running
	// and whose status.version is running.
	createRunningCluster := func(running string, mutate func(*v1beta1.TemporalCluster)) {
		GinkgoHelper()
		datastore := func(name, db string) *v1beta1.DatastoreSpec {
			return &v1beta1.DatastoreSpec{
				Name: name,
				SQL: &v1beta1.SQLSpec{
					User: "temporal", PluginName: "postgres12", DatabaseName: db,
					ConnectAddr: "postgres:5432", ConnectProtocol: "tcp",
				},
				PasswordSecretRef: &v1beta1.SecretKeyReference{Name: "postgres", Key: "PASSWORD"},
			}
		}
		c := &v1beta1.TemporalCluster{
			ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name},
			Spec: v1beta1.TemporalClusterSpec{
				Version:          version.MustNewVersionFromString(running),
				NumHistoryShards: 1,
				Persistence: v1beta1.TemporalPersistenceSpec{
					DefaultStore:    datastore("default", "temporal"),
					VisibilityStore: datastore("visibility", "temporal_visibility"),
				},
				AdminTools: &v1beta1.TemporalAdminToolsSpec{Enabled: true},
			},
		}
		if mutate != nil {
			mutate(c)
		}
		c.Default()
		Expect(k8sClient.Create(ctx, c)).To(Succeed())

		schema := version.MustNewVersionFromString(running)
		c.Status = v1beta1.TemporalClusterStatus{
			Version: running,
			Persistence: &v1beta1.TemporalPersistenceStatus{
				DefaultStore:    &v1beta1.DatastoreStatus{Created: true, Setup: true, Type: v1beta1.PostgresSQL12Datastore, SchemaVersion: schema},
				VisibilityStore: &v1beta1.DatastoreStatus{Created: true, Setup: true, Type: v1beta1.PostgresSQL12Datastore, SchemaVersion: schema},
			},
			Conditions: []metav1.Condition{},
		}
		// status.services is left empty: it is derived from the deployments, which envtest starts
		// without. The first reconcile creates them, and they are not ready until rolled out.
		Expect(k8sClient.Status().Update(ctx, c)).To(Succeed())
	}

	jobs := func() []batchv1.Job {
		GinkgoHelper()
		list := &batchv1.JobList{}
		Expect(k8sClient.List(ctx, list, client.InNamespace(namespace))).To(Succeed())
		return list.Items
	}

	completeJobs := func() {
		GinkgoHelper()
		for _, job := range jobs() {
			if job.Status.Succeeded == 1 || isJobFailed(&job) {
				continue
			}
			now := metav1.Now()
			job.Status.StartTime = &now
			job.Status.CompletionTime = &now
			job.Status.Succeeded = 1
			job.Status.Conditions = []batchv1.JobCondition{
				{Type: batchv1.JobSuccessCriteriaMet, Status: corev1.ConditionTrue, LastTransitionTime: now},
				{Type: batchv1.JobComplete, Status: corev1.ConditionTrue, LastTransitionTime: now},
			}
			Expect(k8sClient.Status().Update(ctx, &job)).To(Succeed())
		}
	}

	failJob := func(name string) {
		GinkgoHelper()
		job := &batchv1.Job{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, job)).To(Succeed())
		now := metav1.Now()
		job.Status.StartTime = &now
		job.Status.Failed = 1
		job.Status.Conditions = []batchv1.JobCondition{
			{Type: batchv1.JobFailureTarget, Status: corev1.ConditionTrue, LastTransitionTime: now, Reason: batchv1.JobReasonBackoffLimitExceeded},
			{Type: batchv1.JobFailed, Status: corev1.ConditionTrue, LastTransitionTime: now, Reason: batchv1.JobReasonBackoffLimitExceeded},
		}
		Expect(k8sClient.Status().Update(ctx, job)).To(Succeed())
	}

	rollOutDeployments := func() {
		GinkgoHelper()
		list := &appsv1.DeploymentList{}
		Expect(k8sClient.List(ctx, list, client.InNamespace(namespace))).To(Succeed())
		for _, d := range list.Items {
			replicas := ptr.Deref(d.Spec.Replicas, 1)
			d.Status = appsv1.DeploymentStatus{
				ObservedGeneration: d.Generation,
				Replicas:           replicas,
				UpdatedReplicas:    replicas,
				ReadyReplicas:      replicas,
				AvailableReplicas:  replicas,
				Conditions: []appsv1.DeploymentCondition{
					{Type: appsv1.DeploymentAvailable, Status: corev1.ConditionTrue, Reason: "MinimumReplicasAvailable"},
					{Type: appsv1.DeploymentProgressing, Status: corev1.ConditionTrue, Reason: "NewReplicaSetAvailable"},
				},
			}
			Expect(k8sClient.Status().Update(ctx, &d)).To(Succeed())
		}
	}

	image := func(service string) string {
		GinkgoHelper()
		d := &appsv1.Deployment{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: key.Name + "-" + service}, d)).To(Succeed())
		return d.Spec.Template.Spec.Containers[0].Image
	}

	currentStep := func() string {
		GinkgoHelper()
		c := getCluster()
		Expect(c.Status.Upgrade).NotTo(BeNil())
		return c.Status.Upgrade.CurrentStep.String()
	}

	// settle reconciles and plays the kube-controller-manager until the cluster reaches the given
	// step with everything done, or fails the test.
	settle := func() {
		GinkgoHelper()
		for range 6 {
			reconcileOnce()
			completeJobs()
			rollOutDeployments()
		}
		reconcileOnce()
	}

	setSpec := func(mutate func(*v1beta1.TemporalCluster)) {
		GinkgoHelper()
		c := getCluster()
		mutate(c)
		Expect(k8sClient.Update(ctx, c)).To(Succeed())
	}

	events := func() string {
		out := []string{}
		for {
			select {
			case e := <-recorder.Events:
				out = append(out, e)
			default:
				return strings.Join(out, "\n")
			}
		}
	}

	It("walks every intermediate minor version, one fully rolled out step at a time", func() {
		createRunningCluster("1.29.7", func(c *v1beta1.TemporalCluster) {
			// A default tag pinned in the manifest, as the docs example does for 1.24.
			c.Spec.AdminTools.Version = "1.29"
		})
		setSpec(func(c *v1beta1.TemporalCluster) {
			c.Spec.Version = version.MustNewVersionFromString("1.32.2")
			c.Spec.Upgrade = &v1beta1.UpgradeSpec{Strategy: v1beta1.AutomaticUpgradeStrategy}
			c.Default()
		})

		// The first reconcile holds the cluster where it runs and confirms it is healthy there.
		reconcileOnce()
		Expect(currentStep()).To(Equal("1.29.7"))
		Expect(image("frontend")).To(Equal("temporalio/server:1.29.7"))
		Expect(jobs()).To(BeEmpty(), "the running version's schema is already in place")

		steps := []string{"1.30.9", "1.31.5", "1.32.2"}
		frontendImages := make([]string, 0, len(steps)+1)
		frontendImages = append(frontendImages, image("frontend"))
		for _, step := range steps {
			rollOutDeployments()
			reconcileOnce()
			Expect(currentStep()).To(Equal(step), "the next step starts once the previous one is healthy")

			// Until the step's schema jobs complete, nothing moves: no deployment is updated and no
			// further step starts, however often the cluster is reconciled.
			for range 3 {
				reconcileOnce()
				Expect(currentStep()).To(Equal(step))
				Expect(image("frontend")).To(Equal(frontendImages[len(frontendImages)-1]))
			}
			stepJobs := []string{}
			for _, j := range jobs() {
				if j.Labels["app.kubernetes.io/version"] == step {
					stepJobs = append(stepJobs, j.Name)
				}
			}
			dashed := strings.ReplaceAll(step, ".", "-")
			Expect(stepJobs).To(ContainElement("cluster-update-default-schema-v-" + dashed))

			completeJobs()
			reconcileOnce() // second schema job
			completeJobs()
			reconcileOnce()
			Expect(image("frontend")).To(Equal("temporalio/server:" + step))
			c := getCluster()
			Expect(c.Status.Persistence.DefaultStore.SchemaVersion.String()).To(Equal(step))
			Expect(c.Status.Persistence.VisibilityStore.SchemaVersion.String()).To(Equal(step))
			Expect(c.Status.Upgrade.Phase).To(Equal(v1beta1.UpgradePhaseProgressing))

			// Deployments updated but not rolled out yet: the step is not complete.
			reconcileOnce()
			Expect(currentStep()).To(Equal(step))
			frontendImages = append(frontendImages, image("frontend"))
		}

		rollOutDeployments()
		reconcileOnce()

		c := getCluster()
		Expect(frontendImages).To(Equal([]string{
			"temporalio/server:1.29.7", "temporalio/server:1.30.9", "temporalio/server:1.31.5", "temporalio/server:1.32.2",
		}))
		Expect(c.Status.Version).To(Equal("1.32.2"))
		Expect(c.Status.Upgrade.Phase).To(Equal(v1beta1.UpgradePhaseCompleted))
		Expect(c.Status.Upgrade.StartVersion.String()).To(Equal("1.29.7"))
		Expect(c.Status.Upgrade.PendingSteps).To(BeEmpty())
		cond := apimeta.FindStatusCondition(c.Status.Conditions, v1beta1.UpgradingCondition)
		Expect(cond).NotTo(BeNil())
		Expect(cond.Status).To(Equal(metav1.ConditionFalse))
		Expect(cond.Reason).To(Equal(v1beta1.UpgradeCompletedReason))
		Expect(apimeta.IsStatusConditionTrue(c.Status.Conditions, v1beta1.ReadyCondition)).To(BeTrue())

		// The spec is never rewritten, and the default-shaped admin tools tag followed each step.
		Expect(c.Spec.Version.String()).To(Equal("1.32.2"))
		Expect(c.Spec.AdminTools.Version).To(Equal("1.29"))
		Expect(image("admintools")).To(Equal("temporalio/admin-tools:1.32"))

		evts := events()
		Expect(evts).To(ContainSubstring("UpgradeStepStarted Upgrading from 1.29.7 to 1.30.9 (target 1.32.2)"))
		Expect(evts).To(ContainSubstring("UpgradeStepStarted Upgrading from 1.31.5 to 1.32.2 (target 1.32.2)"))
		Expect(evts).To(ContainSubstring("UpgradeCompleted Upgrade from 1.29.7 to 1.32.2 completed"))
	})

	It("keeps a custom admin tools tag", func() {
		createRunningCluster("1.31.5", func(c *v1beta1.TemporalCluster) {
			c.Spec.AdminTools.Version = "my-mirror-build"
		})
		setSpec(func(c *v1beta1.TemporalCluster) {
			c.Spec.Version = version.MustNewVersionFromString("1.32.2")
			c.Spec.Upgrade = &v1beta1.UpgradeSpec{Strategy: v1beta1.AutomaticUpgradeStrategy}
			c.Default()
		})
		settle()
		Expect(getCluster().Status.Upgrade.Phase).To(Equal(v1beta1.UpgradePhaseCompleted))
		Expect(image("admintools")).To(Equal("temporalio/admin-tools:my-mirror-build"))
	})

	It("halts on a failed job, stops retrying, and retries once the spec changes", func() {
		createRunningCluster("1.30.9", nil)
		setSpec(func(c *v1beta1.TemporalCluster) {
			c.Spec.Version = version.MustNewVersionFromString("1.32.2")
			c.Spec.Upgrade = &v1beta1.UpgradeSpec{Strategy: v1beta1.AutomaticUpgradeStrategy}
			c.Default()
		})
		reconcileOnce()
		rollOutDeployments()
		reconcileOnce()
		Expect(currentStep()).To(Equal("1.31.5"))
		reconcileOnce()

		failJob("cluster-update-default-schema-v-1-31-5")
		result := reconcileOnce()
		Expect(result).To(Equal(ctrl.Result{}), "a failed step is not requeued")

		c := getCluster()
		Expect(c.Status.Upgrade.Phase).To(Equal(v1beta1.UpgradePhaseFailed))
		Expect(c.Status.Upgrade.FailedGeneration).To(Equal(c.Generation))
		Expect(c.Status.Upgrade.Message).To(ContainSubstring("job cluster-update-default-schema-v-1-31-5 failed"))
		cond := apimeta.FindStatusCondition(c.Status.Conditions, v1beta1.UpgradingCondition)
		Expect(cond.Reason).To(Equal(v1beta1.UpgradeFailedReason))
		Expect(events()).To(ContainSubstring("Warning UpgradeFailed"))

		// Halted: further reconciles change nothing and don't requeue.
		jobsBefore := len(jobs())
		for range 3 {
			Expect(reconcileOnce()).To(Equal(ctrl.Result{}))
		}
		Expect(jobs()).To(HaveLen(jobsBefore))
		Expect(getCluster().Status.Upgrade.Phase).To(Equal(v1beta1.UpgradePhaseFailed))
		Expect(image("frontend")).To(Equal("temporalio/server:1.30.9"))

		// Pausing is a spec change: the failure is cleared and the failed job deleted so it runs again.
		setSpec(func(c *v1beta1.TemporalCluster) { c.Spec.Upgrade.Paused = true })
		reconcileOnce()
		Expect(getCluster().Status.Upgrade.Phase).To(Equal(v1beta1.UpgradePhaseProgressing))
		Expect(events()).To(ContainSubstring("UpgradeRetrying Retrying upgrade step 1.31.5"))
		job := &batchv1.Job{}
		err := k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "cluster-update-default-schema-v-1-31-5"}, job)
		if err == nil {
			// Recreated by the same reconcile: it must be a fresh, unfailed job.
			Expect(isJobFailed(job)).To(BeFalse())
		}

		// The step completes while paused, and the upgrade then holds there.
		settle()
		c = getCluster()
		Expect(c.Status.Version).To(Equal("1.31.5"))
		Expect(c.Status.Upgrade.Phase).To(Equal(v1beta1.UpgradePhasePaused))
		Expect(c.Status.Upgrade.CurrentStep.String()).To(Equal("1.31.5"))
		Expect(c.Status.Upgrade.Message).To(ContainSubstring("next step is 1.32.2"))

		setSpec(func(c *v1beta1.TemporalCluster) { c.Spec.Upgrade.Paused = false })
		settle()
		c = getCluster()
		Expect(c.Status.Version).To(Equal("1.32.2"))
		Expect(c.Status.Upgrade.Phase).To(Equal(v1beta1.UpgradePhaseCompleted))
	})

	It("fails a step that exceeds its timeout", func() {
		createRunningCluster("1.31.5", nil)
		setSpec(func(c *v1beta1.TemporalCluster) {
			c.Spec.Version = version.MustNewVersionFromString("1.32.2")
			c.Spec.Upgrade = &v1beta1.UpgradeSpec{Strategy: v1beta1.AutomaticUpgradeStrategy, StepTimeout: &metav1.Duration{Duration: 10 * time.Minute}}
			c.Default()
		})
		reconcileOnce()
		rollOutDeployments()
		reconcileOnce()
		Expect(currentStep()).To(Equal("1.32.2"))
		result := reconcileOnce()
		Expect(result.RequeueAfter).To(BeNumerically(">", 0), "a step in progress is re-evaluated on a timer")
		Expect(result.RequeueAfter).To(BeNumerically("<=", upgradeProgressRequeue))

		upgradeClock = func() time.Time { return time.Now().Add(11 * time.Minute) }
		Expect(reconcileOnce()).To(Equal(ctrl.Result{}))
		c := getCluster()
		Expect(c.Status.Upgrade.Phase).To(Equal(v1beta1.UpgradePhaseFailed))
		Expect(c.Status.Upgrade.Message).To(ContainSubstring("did not complete within 10m0s (waiting for persistence schema jobs)"))
	})

	It("follows target changes made mid-upgrade", func() {
		createRunningCluster("1.28.4", nil)
		setSpec(func(c *v1beta1.TemporalCluster) {
			c.Spec.Version = version.MustNewVersionFromString("1.30.9")
			c.Spec.Upgrade = &v1beta1.UpgradeSpec{Strategy: v1beta1.AutomaticUpgradeStrategy}
			c.Default()
		})
		reconcileOnce()
		rollOutDeployments()
		reconcileOnce()
		Expect(currentStep()).To(Equal("1.29.7"))
		Expect(getCluster().Status.Upgrade.PendingSteps).To(Equal([]string{"1.30.9"}))

		// Raised to Latest while 1.29.7 is in flight: the step in flight is kept, the plan extends.
		setSpec(func(c *v1beta1.TemporalCluster) { c.Spec.Upgrade.Target = v1beta1.LatestUpgradeTarget })
		reconcileOnce()
		c := getCluster()
		Expect(c.Status.Upgrade.CurrentStep.String()).To(Equal("1.29.7"))
		Expect(c.Status.Upgrade.TargetVersion.String()).To(Equal(version.LatestKnownRelease().String()))
		Expect(c.Status.Upgrade.PendingSteps).To(Equal([]string{"1.30.9", "1.31.5", "1.32.2"}))
		Expect(events()).To(ContainSubstring(fmt.Sprintf("UpgradeTargetChanged Upgrade target changed from 1.30.9 to %s", version.LatestKnownRelease())))

		for range 4 {
			settle()
		}
		c = getCluster()
		Expect(c.Status.Version).To(Equal(version.LatestKnownRelease().String()))
		Expect(c.Status.Upgrade.Phase).To(Equal(v1beta1.UpgradePhaseCompleted))
		Expect(c.Spec.Version.String()).To(Equal("1.30.9"), "the Latest target does not rewrite spec.version")
	})

	It("leaves manual clusters alone and clears the upgrade status when the strategy is dropped", func() {
		createRunningCluster("1.32.2", nil)
		settle()
		c := getCluster()
		Expect(c.Status.Upgrade).To(BeNil())
		Expect(apimeta.FindStatusCondition(c.Status.Conditions, v1beta1.UpgradingCondition)).To(BeNil())

		setSpec(func(c *v1beta1.TemporalCluster) {
			c.Spec.Upgrade = &v1beta1.UpgradeSpec{Strategy: v1beta1.AutomaticUpgradeStrategy}
			c.Default()
		})
		settle()
		Expect(getCluster().Status.Upgrade.Phase).To(Equal(v1beta1.UpgradePhaseCompleted))

		setSpec(func(c *v1beta1.TemporalCluster) { c.Spec.Upgrade = nil })
		reconcileOnce()
		c = getCluster()
		Expect(c.Status.Upgrade).To(BeNil())
		Expect(apimeta.FindStatusCondition(c.Status.Conditions, v1beta1.UpgradingCondition)).To(BeNil())
	})
})

// cacheLikeClient sets the GVK on objects it reads, as the manager's cache-backed client does and a
// direct client does not. The reconciler relies on it: it derives service statuses from the
// Deployments it read back, matching them by GVK.
type cacheLikeClient struct {
	client.Client
}

func (c cacheLikeClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if err := c.Client.Get(ctx, key, obj, opts...); err != nil {
		return err
	}
	gvk, err := apiutil.GVKForObject(obj, c.Scheme())
	if err != nil {
		return err
	}
	obj.GetObjectKind().SetGroupVersionKind(gvk)
	return nil
}
