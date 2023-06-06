# HarborProject Day2 Operations

Harbor Operator is capable of managing the projects of a Harbor instance.

The following operations involving projects are currently supported:

* Create, update and delete projects
* Manage group and user memberships of projects
* Update a projects storage quota

By default, the operator reconciles all `HarborProject` resources every 5 minutes. Changes applied manually to operator-managed projects will be overwritten. The reconciliation interval can be configured using the key `controllers.harborProject.requeueAfterMinutes` in the operator's `values.yaml`.

## The `HarborProject` CustomResourceDefinition

Harbor projects can be managed by deploying a `HarborProject` resource to your Kubernetes cluster.

### `spec`

* `cveAllowList`: List of CVE-strings. This sets the CVE allow list of the project.
* `harborServerConfig`: Name of a `HarborServerConfig` resource containing the reference and configurations for the harbor instance to manage.
* `memberships`: List of members. Members are defined as follows:
  * `name`: Name of the member. Has to match with a existing user or group in the harbor instance.
  * `role`: Role of the member in the project. This controls the member's permissions on the project. Can be either `projectAdmin`, `developer`, `guest` or `maintainer`. See the [Harbor Docs](https://goharbor.io/docs/latest/administration/managing-users/user-permissions-by-role/) for further info on member permissions.
  * `type`: Type of the member, can be `group` or `user`.
* `metadata`: Miscellaneous project metadata.
  * `autoScan`: Boolean. Whether to scan images automatically after pushing.
  * `enableContentTrust`: Boolean. Whether content trust is enabled or not. If enabled, user can't pull unsigned images from this project.
  * `enableContentTrustCosign`: Boolean. Whether cosign content trust is enabled or not. Similar to enableContentTrust, but using cosign.
  * `preventVulnerable`: Boolean. Whether to prevent vulnerable images from running.
  * `public`: Boolean. Whether the project should be public or not.
  * `reuseSysCveAllowlist`: Boolean. Whether this project reuses the system level CVE allowlist for itself. If this is set to `true`, the actual allowlist associated with this project will be ignored.
  * `severity`: If an image's vulnerablilities are higher than the severity defined here, the image can't be pulled. Can be either `none`, `low`, `medium`, `high` or `critical`.
* `projectName`: The name of the harbor project. Has to match harbor's naming rules.
* `storageQuota`: The project's storage quota in human-readable format, like in Kubernetes memory requests/limits (Ti, Gi, Mi, Ki). The Harbor's default value is used if empty.
