# Install with all-in-one deployment manifest

The installation guide documented here help you deploy Harbor operator stack with an all-in-one deployment manifest, which is the recommended way.

## Prerequisites

1. `Kubernetes` cluster (v1.20+) is running (see [Applicative Kubernetes versions](../../README.md#applicative-kubernetes-versions)
   for more information). For local development purpose, check [Kind installation](./kind-installation.md).
1. `cert-manager` (v1.4.4+) is [installed](https://cert-manager.io/docs/installation/kubernetes/).
1. Ingress controller is deployed (see [Ingress controller types](../../README.md#ingress-controller-types) for more information). For default
   ingress controller, check [NGINX ingress controller](https://kubernetes.github.io/ingress-nginx/deploy/) (version should be >1.0).
1. `kubectl` with a proper version(v1.20.1+) is [installed](https://kubernetes.io/docs/tasks/tools/).
1. `kustomize` (optional) with a proper version(v3.8.7+) is [installed](https://kubectl.docs.kubernetes.io/installation/kustomize/).
1. `git` (optional) is [installed](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git).

## One-click installation

Directly apply the all-in-one deployment manifest to your Kubernetes cluster:

```shell
kubectl apply -f https://raw.githubusercontent.com/goharbor/harbor-operator/master/manifests/cluster/deployment.yaml
```

>NOTES: Here we use the deployment manifest in the `main` branch as an example, for deploying the released versions, you can get the deployment manifest in the GitHub release page or find it in the corresponding code branch such as `release-1.3.0`.

Check the installed operators:

```shell
kubectl get pod -n harbor-operator-ns
```

Output:

```log
console-67d5498b88-2hq5d            1/1     Running   0          20m
harbor-operator-54454997d-f6b6g     1/1     Running   0          20m
minio-operator-c4d8f7b4d-h8rwp      1/1     Running   0          20m
postgres-operator-94578ffd5-b4xt7   1/1     Running   0          20m
redisoperator-6b75fc4555-kldnh      1/1     Running   0          20m
```

## Customize deployment manifest

If you want to customize the deployment manifest like editing image settings of operators or [customizing images](../customize-images.md#by-operator-environment-variables) of the deploying Harbor etc., you can clone the code of the specified branch into your computer first.

```shell
git clone https://github.com/goharbor/harbor-operator.git

# Checkout to the specified branch or the specified tag.
# To branch: git checkout <branch-name> e.g.: git checkout release-1.3.0
# To tag: git checkout tags/<tag> -b <branch-name> e.g: git checkout tags/v1.3.0 -b tag-v1.3.0
```

As the resource manifests are not stored in the codebase, then you need to run the following command to generate the related resource manifests before using `kustomize` to build your customized operator deployment manifest:

```shell
make manifests
```

Do necessary modifications to the `manifests/cluster/kustomization.yaml` kustomization template file according to your actual use case and apply the revised deployment manifest to your Kubernetes clusters with command:

```shell
kustomize build manifests/cluster | kubectl apply -f -
```

Of course, generating the updated deployment manifest first and applying it is also ok:

```shell
# Generate
kustomize build manifests/cluster -o customized_deployment.yaml

# Apply

kubectl apply -f customized_deployment.yaml
```

>NOTES: For editing operator images, you can also try command `kustomize edit set image goharbor/harbor-operator=ns/my-operator:mytag` under kustomization folder 'manifests/cluster'.

## Delete operator

Delete the harbor operator stack by the deployment manifest:

```shell
kubectl delete -f https://raw.githubusercontent.com/goharbor/harbor-operator/master/manifests/cluster/deployment.yaml
```

Or delete the harbor operator stack by the kustomization template:

```shell
kustomize build manifests/cluster | kubectl delete -f -
```

## What's next

If the Harbor operator is successfully installed, you can follow the guide
shown [here](../tutorial.md#deploy-harbor-cluster) to deploy your Harbor cluster to your Kubernetes cluster and try it.
