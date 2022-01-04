# Day2 Operation

> NOTES: So far, the day2 related operations are not formally contained in Harbor operator releases yet.

The day2 operations includes:

* Mapping k8s namespace and harbor project
* Pulling secret auto injection
* Image path auto rewriting
* Transparent proxy cache
* Apply configuration changes

For more info, check [Here](https://github.com/goharbor/harbor-operator/docs/configurations/day2-config.md).

## Overall design

The diagram shown below describes the overall design of the day 2 operation features.

![day2 operation arch design](../images/arch/day2-operator-design.jpg)

### Refer the deployed Harbor

A cluster scoped CRD `HarborServerConfiguration` (HSC) is introduced to hold the accessing info like 'host' and 'credential' of the target Harbor registry to which the day2 operations will be applied. The target Harbor registry can be either being deployed inside or installed outside the Kubernetes cluster. It only needs to make sure the target Harbor registry is accessible for the related operator controllers. A `HarborServerConfiguration` CR can be marked as default and used when there is no specified settings set. Additionally, a container image path rewriting rule list can be appended to the `HarborServerConfiguration` CR for being used in the image path rewriting scenario. e.g:

```yaml
spec:
  rules:
    - registryRegex: "^docker.io$"
      project: myHarborProject
```

The `HarborServerConfiguration` controller will regularly update the status of the CR to correctly reflect the health of the associated Harbor registry.

### Watch namespace

The namespace controller of Harbor operator regularly watches the Kubernetes namespace changes and create or remove the `PullSecretBinding` CR depending on whether the annotation `goharbor.io/secret-issuer` that refers to an existing 'HSC' CR is applied to the namespace or not.

The optional annotation `goharbor.io/project` is used to specify which existing harbor project mapping to the Kubernetes namespace.
If it is not set, same name with Kubernetes namespace is chosen.

The optional annotation `goharbor.io/robot` needs to be set with an ID of the existing Harbor robot account in the project specified by the above annotation `goharbor.io/project` if the annotation `goharbor.io/project` is set. This restriction can make sure the user who sets the project has the proper permissions to that project.

The optional annotation `goharbor.io/service-account` can be used to specify which service account should be referred to bind the pull secret. If it is not set, then the default service account of the namespace is chosen.

### Populate mappings and inject pull secret

The `PullSecretBinding` (PSB) CR is designed to maintaining the mappings between the Kubernetes namespace ,and the Harbor project as well as the image pulling secret ,and the Harbor robot account. Generally, it is handled by the namespace controller of Harbor operator.

The `PullSecretBinding` CR controller is in charge of reconciling the 'PSB' CR. It will assure the project specified in the annotation `goharbor.io/project` (or the default mapping one) and the robot account specified in the annotation
`goharbor.io/robot` (or new one) existing in the Harbor registry referred by the 'HSC'. The controller wraps the mapping robot account as Kubernetes secret and inject this secret into the service account referred with annotation `goharbor.io/service-account`.

### Image path rewriting

A mutating webhook is created to help rewrite the container image path to point to the mapping project of the Harbor registry referred by the 'HSC'. If the 'HSC' annotated by `goharbor.io/secret-issuer` exists, it will be used. Otherwise, the default 'HSC' will be chosen. However, if the default 'HSC' does not exist, no rewriting action will happen.

Under the case of using the 'HSC' annotated by `goharbor.io/secret-issuer`, an extra annotation `goharbor.io/image-rewrite=auto` should be appended to the namespace.

Under the case of having default 'HSC', to disable the image rewriting, an extra annotation `goharbor.io/image-rewrite=disabled`
should be appended to the namespace.

### Configuration

Set the Harbor system settings via a configuration CRD. At this time, it is not supported anywhere, but a configMap based configuration has been supported in the released Harbor operator. For how to use it, check [day2 config](../day2/day2-operations.md).

## References

* Learn more, check Harbor day2 operator [prototype project](https://github.com/goharbor/harbor-operator).
