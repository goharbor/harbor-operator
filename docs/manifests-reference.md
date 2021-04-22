# Manifests reference

There are some manifest files located under the [./manifests](../manifests) folder. Their usages are documented here.

## Usages of manifest files

### manifests/cluster

This folder contains the kustomization templates ,and the deployment manifest yaml for deploying the full Harbor operator stack (including Harbor operator as well as all other operators of the dependent services such as PostgreSQL, Redis and Minio) in an all-in-one way.

Use the kustomization template by the following way:

```shell
kustomize build manifests/cluster | kubectl apply -f -

# OR
kustomize build manifests/cluster | kubectl delete -f -

# OR

kustomize build manifests/cluster -o customzied_deployment.yaml
```

Use the deployment manifest yaml by the following way:

```shell
kubectl apply -f manifests/cluster/deployment.yaml

# OR
kubectl delete -f manifests/cluster/deployment.yaml
```

For more info, check [kustomization-all-in-one](./installation/kustomization-all-in-one.md).

### manifests/harbor

This folder contains the kustomization templates ,and the deployment manifest yaml for deploying the Harbor operator itself (no operators of the dependent services such as PostgreSQL, Redis and Minio).

Use the kustomization template by the following way:

```shell
kustomize build manifests/harbor | kubectl apply -f -

# OR
kustomize build manifests/harbor | kubectl delete -f -

# OR

kustomize build manifests/harbor -o customzied_deployment.yaml
```

Use the deployment manifest yaml by the following way:

```shell
kubectl apply -f manifests/harbor/deployment.yaml

# OR
kubectl delete -f manifests/harbor/deployment.yaml
```

For more info, check [kustomization-custom](./installation/kustomization-custom.md).

### manifests/samples

This folder contains several sample manifests for you to deploy Harbor cluster with different structures.

|  Manifests   |   Description    |
|--------------|------------------|
| [harbor_core_fs.yaml](../manifests/samples/harbor_core_fs.yaml) |Deploy the Harbor cluster with the structure: harbor core components + filesystem storage(PV) + in-cluster PostgreSQL + in-cluster Redis |
| [harbor_core_minio.yaml](../manifests/samples/harbor_core_minio.yaml) |Deploy the Harbor cluster with the structure: harbor core components + in-cluster storage(Minio) + in-cluster PostgreSQL + in-cluster Redis |
| [harbor_full_fs.yaml](../manifests/samples/harbor_full_fs.yaml) |Deploy the Harbor cluster with the structure: harbor all components + filesystem storage(PV) + in-cluster PostgreSQL + in-cluster Redis |
| [harbor_full_minio.yaml](../manifests/samples/harbor_full_minio.yaml) |Deploy the Harbor cluster with the structure: harbor all components + in-cluster storage(Minio) + in-cluster PostgreSQL + in-cluster Redis |
| [harborcluster_standard_stack_external.yaml](../manifests/samples/harborcluster_standard_stack_external.yaml) |Deploy the Harbor cluster with the structure: harbor all components + filesystem storage(PV) + external PostgreSQL + external Redis|

> NOTE: `external` means you need to pre-deploy the required services; `in-cluster` means the Harbor operator will create the required services while deploying the Harbor cluster.

## What's next

Follow the [tutorial](./tutorial.md) guideline to install the Harbor operator and deploy Harbor cluster to your cluster.
