# Pinning image tags and digests

By default the operator derives every container image tag from the configured
`version`: the server, web UI and admin tools images are all deployed as
`<image>:<version>`. This is convenient, but it means:

- the deployed tag is coupled to `version`, even though `version` also drives
  the persistence schema migrations, and
- a mutable tag can be repointed upstream, so the running image can change
  without any change to the `TemporalCluster` resource.

Each component (`spec`, `spec.ui`, `spec.admintools`) therefore exposes two
optional fields:

- **`imageTag`** — overrides the tag used for the image. When set, it is used
  instead of `version`. For the server, `version` still drives the persistence
  schema version regardless of this override — only the deployed image tag
  changes. This is useful when an image is published with a tag that differs
  from the Temporal version (for example admin-tools images).
- **`imageDigest`** — pins the image to an immutable digest (`sha256:...`). When
  set, it is appended to the reference as `<image>:<tag>@<digest>`. The tag is
  kept for readability while the digest guarantees the running image cannot
  change behind a mutable tag. This is recommended for security-sensitive or
  compliance-driven deployments (e.g. supply-chain / image-provenance
  requirements).

Both fields are optional and can be combined. The resulting image reference is:

| `imageTag` | `imageDigest` | Resulting reference          |
| ---------- | ------------- | ---------------------------- |
| unset      | unset         | `image:version`              |
| set        | unset         | `image:imageTag`             |
| unset      | set           | `image:version@digest`       |
| set        | set           | `image:imageTag@digest`      |

## Example: pin the server and UI to a digest

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
spec:
  version: 1.28.2
  image: temporalio/server
  imageDigest: sha256:459aa877ef82c9a7f0ce688f2451ec8e576142c4efefecb2988d2f5f8b12699e
  ui:
    enabled: true
    version: 2.44.1
    image: temporalio/ui
    imageDigest: sha256:0000000000000000000000000000000000000000000000000000000000000000
```

## Example: override the admin-tools tag

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
spec:
  version: 1.28.2
  admintools:
    enabled: true
    image: temporalio/admin-tools
    # the admin-tools image for this version is published under a different tag
    imageTag: 1.28.2-tctl-1.18.1-cli-0.13.2
```
