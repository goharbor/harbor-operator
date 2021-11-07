# Day2 Operation

Day2 Operation provides the following features to help users get better experiences of using [Harbor](https://github.com/goharbor/harbor)
or apply some Day2 operations to the Harbor deployed in the Kubernetes cluster:

- [x] **Mapping k8s namespace and harbor project**: make sure there is a relevant project existing at linked Harbor side for the
 specified k8s namespace pulling image from there (bind specified one or create new).
- [x] **Pulling secret auto injection**: auto create robot account in the corresponding Harbor project and bind it to the
 related service account of the related k8s namespace to avoid explicitly specifying image pulling secret in the
 deployment manifest yaml.
- [x] **image path auto rewriting**: rewrite the pulling path of the matched workload images (e.g: no full repository path specified)
 being deployed in the specified k8s namespace to the corresponding project at the linked Harbor.
- [x] **transparent proxy cache**: rewrite the pulling path of the matched workload images to the proxy cache project of the linked Harbor.
- [ ] apply configuration changes: update the system configurations of the linked Harbor with Kubernetes way by providing a configMap.
- [ ] certificate population: populate the CA of the linked Harbor instance to the container runtimes of all the cluster workers and let workers trust it to avoid image pulling issues.
- [ ] TBD

## Overall Design

The diagram below shows the overall design of this project:

![overall design](./docs/assets/4k8s-automation.png)

### Feature

#### Image rewrite

Image rewrite will rewrite the original image paths to the specified Harbor projects by following the pre-defined rewriting rules via Kubernetes admission control webhook.

A cluster scoped Kubernetes CR named `HarborServerConfiguration` is designed to keep the Harbor server access info by providing the access
host and access key & secret (key and secret should be wrapped into a kubernetes secret) for future referring.
It has a `default` field to define whether this HSC will be applied to all namespaces. There could be only one default HSC.

Rewriting rule is a k-v pair to specify images from which repo should be redirected to which harbor project:
`"docker.io": "harborproject1"` or `"*": "harborproejct2"`

Here we should pay attention is the key "*" means images from any repo are redirected to harbor project "harborproejct2".

Rewriting rules will be defined as a rule list like:

```shell
rules:
  - docker.io:harborproject1
  - *:harborproejct2
  - quay.io:harborproejct3
```

**Definition location:**

The rewriting rules can be defined into two places, one is in the HSC spec and another is in a configMap.

Rules in HSC spec are for the whole cluster scope. The rules will be applied to the namespaces selected by the namespace selector of HSC.

The rules defined in the configMap are only visible to the namespace where the configMap is living. Use annotation of namespace `goharbor.io/image-rewrite`(rename to `goharbor.io/rewriting-rules`) to link the rule configMap.

**The priority:**

Rules in configMap > rules in HSC referenced by ConfigMap > default HSC > "*" rule

For example. images from `docker.io` will be rewritten to 'harborproject1' and images from `quay.io` will be rewritten to 'harborproejct3'. The images from `gcr.io` or `ghcr.io` will both be rewritten to 'harborproejct2' by following the "*" rule.

**Assumptions:**

Only 1 HSC as default. (ctrl has to make sure this constraint)
Default HSC is appliable for all namespaces as the default behavior (except its namespace selector is configured).
HSC can have a namespace selector to specify its influencing scope.
Namespace selector is optional. The empty selector means adapting all.

Namespace admin can create a configMap to customize image rewriting for the specified namespace:

**Content of configMap:**

```yaml
hsc: myHscName ## if this ns missing the selector of the specfying HSC, log warnning and no action.
rewriting: on ## or off
rules: -|
  - docker.io:harborproject1-1
  - *:harborproejct2-1
  - quay.io:harborproejct3-1
```

Add annotation `goharbor.io/rewriting-rules=configMapName` to the namespace to enable the rewriting.

Merging rules: rules defined in configMap has higher priority if conflicts happened.

#### Project mapping and secret injection

When doing project mapping and secret injection, an annotation `goharbor.io/project` MUST be added to the specified namespace ( if `goharbor.io/project` is
not set, that means the mapping/injection function is not enabled).

A CR `PullSecretBinding` is created to keep the relationship between Kubernetes resources and Harbor resources.

- the mapping project is recorded in annotation `annotation:goharbor.io/project` of the CR `PullSecretBinding`.
- the linked robot account is recorded in annotation `annotation:goharbor.io/robot` of the CR `PullSecretBinding`.
- make sure the linked robot account is wrapped as a Kubernetes secret and bind with the service account that is specified in the annotation `annotation:goharbor.io/service-account` of the namespace.

- If `goharbor.io/project`=*, then check whether annotation `goharbor.io/secret-issuer` (it should be renamed to `goharbor.io/harbor`) which is pointing to an HSC (It should be provided by the cluster-admin) is set or not. If it is not set, then back off to the default HSC. If there is no default HSC, then an error should be raised. When HSC is ready, create a Harbor project with the same name of the namespace in that HSC and also create a robot account in the newly created Harbor project.  After the robot account is created, the identity of the created robot account should be recorded into the annotation `goharbro.io/robot` (it should be renamed to `goharbor.io/robot-account`).
- If `goharbor.io/project=<project name>`, then the annotation `goharbor.io/robot=<robot_ID>` MUST also be set to a valid robot account that is living in the project representing by the annotation `goharbor.io/project`. The controller has to make sure the robot specified in the annotation can be used to access the project (by accessing API with that robot account).
  - If the specified project name does not exist or the robot account provided does not mismatch, then log the error (should not return an error in the reconcile process)

Then a PSB can be created to track the relationship and bind pull secret to the service account:

- wrap the robot account as a secret
- bind the secret to the service account which is specified by the annotation `goharbor.io/service-account` (this is optional, if it is not set, then use the default service account under the namespace)

## Installation

For trying this project, you can follow the guideline shown below to quickly install the operator and webhook to your cluster:
(tools like `git`,`make`,`kustomize` and `kubectl` should be installed and available in the $PATH)

1. Clone the repository

1. Build the image (no official image built so far)

```shell script
make docker-build && make docker-push
```

1. Deploy the operator to the Kubernetes cluster

```shell script
make deploy
```

1. Check the status of the operator

```shell script
kubectl get all -n harbor-operator-ns
```

1. Uninstall the operator

```shell script
kustomize build config/default | kubectl delete -f -
```

## Usages

### HarborServerConfiguration CR

Register your Harbor in a `HarborServerConfiguration` CR:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mysecret
  namespace: kube-system
type: Opaque
data:
  accessKey: YWRtaW4=
  accessSecret: SGFyYm9yMTIzNDU=
---
apiVersion: goharbor.io/v1alpha1
kind: HarborServerConfiguration
metadata:
  name: harborserverconfiguration-sample
spec:
  default: true ## whether it will be default global hsc
  serverURL: 10.168.167.189
  accessCredential:
    namespace: kube-system
    accessSecretRef: mysecret
  version: 2.1.0
  inSecure: true
  rules: ## rules to define to rewrite image path
  - "docker.io,myharbor"    ## <repo-regex>,<harbor-project>
  namespaceSelector:
    matchLabels:
      usethisHSC: true
```

Create it:

```shell script
kubectl apply -f hsc.yaml
```

Use the following command to check the `HarborServerConfiguration` CR (short name: `hsc`):

```shell script
kubectl get hsc
```

### Pulling secret injection

Add related annotations to your namespace when enabling secret injection:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: sz-namespace1
  annotations:
    goharbor.io/harbor: harborserverconfiguration-sample
    goharbor.io/service-account: default
    goharbor.io/project: "*"
```

Create it:

```shell script
kubectl apply -f namespace.yaml
```

After the automation is completed, a CR `PullSecretBinding` is created:

```shell script
kubectl get psb -n sz-namespace1

# output
#NAME             HARBOR SERVER                      SERVICE ACCOUNT   STATUS
#binding-txushc   harborserverconfiguration-sample   default           ready
```

Get the details of the psb/binding-xxx:

```shell script
k8s get psb/binding-txushc -n sz-namespace1 -o yaml
```

Output details:

```yaml
apiVersion: goharbor.io/v1alpha1
kind: PullSecretBinding
metadata:
  annotations:
    goharbor.io/project: sz-namespace1-axtnd8
    goharbor.io/robot: "31"
    goharbor.io/robot-secret: regsecret-sab3pq
  creationTimestamp: "2020-12-02T15:21:48Z"
  finalizers:
  - psb.finalizers.resource.goharbor.io
  generation: 1
  name: binding-txushc
  namespace: sz-namespace1
  ownerReferences:
  - apiVersion: v1
    blockOwnerDeletion: true
    controller: true
    kind: Namespace
    name: sz-namespace1
    uid: 810efadd-b560-4791-8007-8decaf2fbb1c
  resourceVersion: "2500851"
  selfLink: /apis/goharbor.io/v1alpha1/namespaces/sz-namespace1/pullsecretbindings/binding-txushc
  uid: f5b4f68a-4657-4f89-b231-0fc96c03ca00
spec:
  harborServerConfig: harborserverconfiguration-sample
  serviceAccount: default
status:
  conditions: []
  status: ready
```

The related auto-generated data is recorded in the related annotations:

```yaml
annotations:
  goharbor.io/project: sz-namespace1-axtnd8
  goharbor.io/robot: "31"
  goharbor.io/robot-secret: regsecret-sab3pq
```

### Image path rewrite

To enable image rewrite, set the rules section in hsc, or set annotation to refer to a configMap that contains rules and hsc

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: sz-namespace1
  annotations:
    goharbor.io/harbor: harborserverconfiguration-sample
    goharbor.io/service-account: default
    goharbor.io/rewriting-rules: sz-namespace1
```

Corresponding ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: sz-namespace1
  namespace: sz-namespace1
data:
  hsc: harbor2
  rewriting: "on"
  rules: | # configMap doesn't support storing nested string
    docker.io,highestproject
    gcr.io,a

```

Corresponding HSC

```yaml
apiVersion: goharbor.io/v1alpha1
kind: HarborServerConfiguration
metadata:
  name: harborserverconfiguration-sample
spec:
  serverURL: 10.168.167.12
  accessCredential:
    namespace: kube-system
    accessSecretRef: mysecret
  version: 2.1.0
  inSecure: true
  rules: ## rules to define to rewrite image path
  - "docker.io,testharbor"    ## <repo-regex>,<harbor-project>

```

As mentioned before, the mutating webhook will rewrite all the images of the deploying pods which has no registry host
prefix to the flowing pattern:

`image:tag => <hsc/hsc-name.[spec.serverURL]>/<psb/binding-xxx.[metadata.annotations[goharbor.io/project]]>/image:tag`
