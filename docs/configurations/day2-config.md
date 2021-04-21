# Day2 configurations

## Step 1

First you need to prepare a config map to provide your harbor configurations, apply the config map in the same namespace as harborcluster. In particular, you need to add an annotation to your config map to mark which harborcluster it is acting on.

In addition, in order to protect the password from being displayed directly in the config map, you need to define the password inside the secret, and then specify the name of the secret in the configuration. We currently offer these type of secret configurations fields: `"email_password", "ldap_search_password", "uaa_client_secret", "oidc_client_secret"`.

The configuration items can be found in [harbor swagger](https://github.com/goharbor/harbor/blob/c701ce09fa2a34b7ea26addb431154f9812de8af/api/v2.0/legacy_swagger.yaml#L516).

### Example

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: secret-sample
  namespace: cluster-sample-ns
type: Opaque
data:
  # the key is same with fields name.
  email_password: YmFyCg==
```

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  # namespace same with harborcluster cr namespace.
  namespace: cluster-sample-ns
  annotations:
    # required.
    # if not define the anno, the config map will not work.
    # the key is `goharbor.io/configuration`, and the value is your harborcluster cr name.
    goharbor.io/configuration: harborcluster-sample
data:
  # provide your harbor configuration by yaml format.
  config.yaml: |
    email_ssl: true
    email_password: secret-sample # the value is the name of secret which store the email_password.
```

## Step 2

Just apply the config map to your kubernetes cluster, if you have related secrets, please apply secrets before config map.  The result of configurate harbor will store in the harborcluster status (with `ConfigurationReady` condition).

```yaml
status:
  conditions:
  - lastTransitionTime: "2021-04-20T10:49:12Z"
    status: "True"
    type: StorageReady
  - lastTransitionTime: "2021-04-20T10:50:23Z"
    message: Harbor component database secrets are already create
    reason: Database is ready
    status: "True"
    type: DatabaseReady
  - lastTransitionTime: "2021-04-20T10:49:13Z"
    message: harbor component redis secrets are already create.
    reason: redis already ready
    status: "True"
    type: CacheReady
  - message: New generation detected
    reason: newGeneration
    status: "True"
    type: InProgress
  - lastTransitionTime: "2021-04-20T10:51:40Z"
    message: JobService.goharbor.io cluster-sample-ns/harborcluster-sample-harbor-harbor
    reason: dependencyStatus
    status: "False"
    type: ServiceReady
  - lastTransitionTime: "2021-04-20T12:54:36Z"
    message: harbor configuraion has been applied successfully.
    reason: ConfigurationApplySuccess
    status: "True"
    type: ConfigurationReady
```
