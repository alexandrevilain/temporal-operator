# Upgrades

Temporal Server only supports upgrading one minor version at a time (`1.n.x` to `1.n+1.x`):
each minor release may migrate the persistence schema, and those migrations build on each other.
The operator enforces this rule through its validating webhook.

There are two upgrade strategies, configured through `spec.upgrade.strategy`:

- `Manual` (the default, also used when `spec.upgrade` is not set): you move `spec.version`
  one minor version at a time and wait for each rollout to finish before the next change.
- `Automatic`: you set the version you want to reach, and the operator rolls out every
  intermediate minor version for you, one at a time.

## Automatic upgrades

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
  namespace: demo
spec:
  # The version the cluster converges to.
  version: 1.32.2
  upgrade:
    strategy: Automatic
    # Optional, defaults to "Version" (converge to spec.version).
    target: Version
    # Optional, defaults to 1h.
    stepTimeout: 1h
    # Optional: set to true to hold the cluster before the next step.
    paused: false
  # [...]
```

With `strategy: Automatic`, `spec.version` may be set to any later supported version, for example
from `1.24.3` straight to `1.32.2`. The operator then:

1. Confirms the cluster is healthy at the version it runs.
2. Picks the next step: the newest known, non-broken patch release of the next minor version
   (`1.25.2`, then `1.26.3`, `1.27.4`, ...). The last step is `spec.version` itself.
3. Runs the step's schema update jobs, then rolls out every Temporal service at the step's version.
4. Waits until the step is complete: schema jobs succeeded, `status.version` equals the step,
   and every service deployment is ready at that version.
5. Repeats from 2 until the cluster runs the target.

`spec.version` is never rewritten by the operator, so the manifest you apply (with `kubectl apply`,
Argo CD, Flux, ...) stays the source of truth.

The patch releases used for intermediate steps come from a release catalog built into the
operator, as the operator does not query registries at runtime. Releases the operator marks as
broken (for example `1.26.0`, `1.26.1`, `1.27.0`) are never used. If you pull images from a mirror,
make sure it holds the `temporalio/server` and `temporalio/admin-tools` tags of every intermediate
step. You can list them in advance from `status.upgrade.pendingSteps` once the upgrade starts, or
check the release catalog in `pkg/version/releases.go`.

### Following the newest release: `target: Latest`

```yaml
spec:
  version: 1.32.2
  upgrade:
    strategy: Automatic
    target: Latest
```

With `target: Latest` the cluster converges to the newest Temporal release known to the running
operator, whatever `spec.version` says. **Upgrading the operator to a release that knows a newer
Temporal version therefore starts a new, possibly multi-step, Temporal upgrade.** The webhook warns
about this whenever the field is set. To go back to `target: Version` (or to the `Manual` strategy),
first set `spec.version` to the version the cluster reached.

### Progress and status

The upgrade is reported in `status.upgrade` and through the `Upgrading` condition:

```yaml
status:
  version: 1.26.3
  upgrade:
    phase: Progressing      # Progressing, Paused, Failed or Completed
    startVersion: 1.24.3
    targetVersion: 1.32.2
    currentStep: 1.27.4
    pendingSteps: [1.28.4, 1.29.7, 1.30.9, 1.31.5, 1.32.2]
    stepStartTime: "2026-11-10T01:02:03Z"
    message: Rolling out 1.27.4, target 1.32.2 (waiting for services: history)
  conditions:
    - type: Upgrading
      status: "True"
      reason: UpgradeInProgress
```

The operator also emits events: `UpgradeStepStarted`, `UpgradeCompleted`, `UpgradeTargetChanged`,
`UpgradeRetrying` and `UpgradeFailed` (a Warning).

```shell
kubectl get temporalcluster prod -o jsonpath='{.status.upgrade}'
kubectl describe temporalcluster prod   # events
```

### Failures

A step fails when one of its jobs fails, or when it doesn't complete within `spec.upgrade.stepTimeout`.
When a step fails, the operator:

- sets `status.upgrade.phase: Failed` and the `Upgrading` condition to `False` with reason
  `UpgradeFailed`, and explains what went wrong in the message;
- emits a `UpgradeFailed` warning event;
- **stops reconciling the cluster's workloads and jobs** and does not requeue the cluster. The
  cluster is left exactly as it was, nothing is retried behind your back, and no further step starts.

The schema can't be rolled back, so the operator never moves a failed upgrade back to an older
version. To retry, fix the cause (look at the failed job's logs, or at the pods of the services
listed in the message), then make any change to the spec. Toggling `spec.upgrade.paused` is the
simplest. The operator then deletes the step's failed jobs so they run again, and restarts the
step timeout.

### Pausing

`spec.upgrade.paused: true` lets the step in progress finish, then holds the cluster
(`phase: Paused`). The message names the next step. Set it back to `false` to continue.

### Changing the target during an upgrade

- **Raising** the target (or switching to `target: Latest`) keeps the step in flight and extends
  the plan.
- **Lowering** the target is accepted as long as it is not below `status.upgrade.currentStep`
  or anything else the cluster has reached (its running version, a service already rolled out, a
  schema already migrated). The webhook rejects anything lower, because that would be a downgrade.
- **Leaving** the automatic strategy (setting `strategy: Manual` or removing `spec.upgrade`) is only
  accepted once `spec.version` is at most one minor version ahead of what the cluster has reached.
  Otherwise the cluster would jump several minor versions at once. Pause the upgrade instead if you
  only want to stop it.

### Validation at every step

Some settings depend on the server version, for example `sql.passwordCommand` needs Temporal
`>= 1.31`. With automatic upgrades the cluster renders every intermediate version, so the
webhook validates the spec at each step and reports, for example,
`at upgrade step 1.30.9: sql.passwordCommand requires Temporal >= 1.31.0.`. Add such settings once
the cluster has reached the version that supports them.

### Admin tools and UI

The admin tools image tag (`spec.admintools.version`) follows each step when it is unset or is one
of the operator's default tags. That includes tags copied from these docs, such as
`1.24.2-tctl-1.18.1-cli-1.0.0`, or a `1.29` style tag. It is resolved for every step as the
operator would default it: `1.25`, `1.26`, ... A custom tag (for example a mirror-specific
build) is used unchanged at every step, and the webhook warns about it.

The schema jobs always use the admin tools tag for the step's version, so schema migrations are
not affected by a pinned tag.

The Temporal UI (`spec.ui.version`) is not changed by automatic upgrades: the UI is released
independently of the server. Bump it yourself once the upgrade has completed.

### Example: upgrading a cluster from 1.24 to 1.32

```shell
kubectl -n demo patch temporalcluster prod --type=merge \
  -p '{"spec":{"version":"1.32.2","upgrade":{"strategy":"Automatic"}}}'
# Watch progress:
kubectl -n demo get temporalcluster prod -w \
  -o custom-columns=NAME:.metadata.name,VERSION:.status.version,STEP:.status.upgrade.currentStep,PHASE:.status.upgrade.phase
```

Before upgrading a production cluster, review the Temporal release notes of every minor version in
between, back up the persistence databases, and consider trying the same path on a disposable
cluster first.
