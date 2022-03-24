# Day2 configurations

Initially, we configure harbor by means of configmap, but currently we recommend using `HarborConfiguration` CRD to configure harbor, for the configmap method will be deprecated in version 1.2, those who have used and still use configmap will be automatically converted to `HarborConfiguration` CR by the controller, and automatically remove old configmap.
> The harbor configuration items can be found in [harbor swagger](https://github.com/goharbor/harbor/blob/0867a6bfd6f33149f86a7ae8a740f5e1f976cafa/api/v2.0/swagger.yaml#L7990).

## ConfigMap (deprecated)

First you need to prepare a config map to provide your harbor configurations, apply the config map in the same namespace as harborcluster. In particular, you need to add an annotation (`goharbor.io/configuration: <harbor cluster name>`) to your config map to mark which harborcluster it is acting on.

In addition, in order to protect the password from being displayed directly in the config map, you need to define the password inside the secret, and then specify the name of the secret in the configuration. We currently offer these type of secret configurations fields: `"email_password", "ldap_search_password", "uaa_client_secret", "oidc_client_secret"`.

**ConfigMap example**:

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

## CRD-based HarborConfiguration

**Example of HarborConfiguration**:

```yaml
apiVersion: goharbor.io/v1beta1
kind: HarborConfiguration
metadata:
  name: test-config
  namespace: cluster-sample-ns
spec:
  # your harbor configuration
  configuration:
    robotTokenDuration: 45
    robotNamePrefix: harbor$
    notificationEnable: false
  harborClusterRef: harborcluster-sample
```

After apply your `HarborConfiguration` CR to kubernetes cluster, the controller of `HarborConfiguration` will apply your configuration to harbor instance, you can see the result of configuration from CR status.

```yaml
status:
  lastApplyTime: "2021-06-04T06:07:53Z"
  lastConfiguration:
    configuration:
      robotTokenDuration: 45
      robotNamePrefix: harbor$
      notificationEnable: false
  status: Success
```
