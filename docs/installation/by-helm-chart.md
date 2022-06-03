# Install Harbor operator by helm chart

The Harbor operator can also be deployed via the operator helm chart provided. What needs to be reminded here is the operator helm chart only deploys Harbor operator itself and does not cover any operator installations of the dependent services (such as PostgreSQL, Redis and Storage(Minio)).

## Prerequisites

1. `Kubernetes` cluster (v1.20+) is running (see [Applicative Kubernetes versions](../../README.md#applicative-kubernetes-versions)
   for more information). For local development purpose, check [Kind installation](./kind-installation.md).
1. `cert-manager` (v1.4.4+) is [installed](https://cert-manager.io/docs/installation/kubernetes/).
1. Ingress controller is deployed (see [Ingress controller types](../../README.md#ingress-controller-types) for more information). For default
   ingress controller, check [NGINX ingress controller](https://kubernetes.github.io/ingress-nginx/deploy/) (version should be >1.0).
1. `kubectl` with a proper version(v1.20.1+) is [installed](https://kubernetes.io/docs/tasks/tools/).
1. `git` (optional) is [installed](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git).
1. `make` and `curl` (optional) are installed.
1. `helm` v3 is [installed](https://helm.sh/docs/intro/install/).

## Get Helm chart

There are several ways to get the Harbor operator helm chart:

1. From the public helm chart repository

    ```shell
    helm repo add stable <PLACEHODLER>
    ```

1. Download it from the Harbor operator release page

    ```shell
    curl -sL -o harbor-operator-x.y.z-build.tgz <PLACEHOLDER>
    ```

1. Generate from the codebase

    ```shell
    git clone https://github.com/goharbor/harbor-operator.git && \
    cd harbor-operator && \
    git checkout release-1.3.0 && \
    make helm-generate RELEASE_VERSION=v1.3.0

    # Checkout to the specified branch or the specified tag
    # To branch: git checkout <branch-name> e.g.: git checkout release-1.3.0
    # To tag: git checkout tags/<tag> -b <branch-name> e.g: git checkout tags/v1.3.0 -b tag-v1.3.0

    # chart is generated to `charts/harbor-operator-v1.3.0.tgz`
    ```

## Deploy Harbor operator with chart

> Under the restriction of helm chart upgrades, CRDs should be updated manually before the deployment of the Harbor operator when the older version(`<v1.3.0`) helm chart has been deployed.

Run the `helm` command to install the harbor operator to your cluster:

```shell
# Change chart path depends on how do you get the helm chart.
helm upgrade --namespace harbor-operator-ns --install harbor-operator charts/harbor-operator-v1.3.0.tgz --set-string image.repository=ghcr.io/goharbor/harbor-operator --set-string image.tag=v1.3.0
```

For what settings you can override with `--set`, `--set-string`, `--set-file` or `--values`, you can refer to the [values.yaml](../../charts/harbor-operator/values.yaml) file.

Once the installation is finished you can check the installation status with either `helm` or `kubectl`.

With `helm`:

```shell
helm status harbor-operator --namespace harbor-operator-ns
```

The command will output related info the installed release:

```log
Release "harbor-operator" does not exist. Installing it now.
NAME: harbor-operator
LAST DEPLOYED: Fri Apr 16 02:56:21 2021
NAMESPACE: harbor-operator-ns
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
1. Get the application URL by running these commands:
  export POD_NAME=$(kubectl get pods --namespace harbor-operator-ns -l "app.kubernetes.io/name=harbor-operator,app.kubernetes.io/instance=harbor-operator" -o jsonpath="{.items[0].metadata.name}")
  echo "Visit http://127.0.0.1:8080 to use your application"
  kubectl --namespace harbor-operator-ns port-forward $POD_NAME 8080:80
```

With `kubectl`:

```shell
kubectl get po -n harbor-operator-ns
```

Output:

```log
NAME                                              READY   STATUS    RESTARTS   AGE
harbor-operator-harbor-operator-865687669-bqnb5   1/1     Running   0          24m
```

## Delete the harbor operator

Run the following `helm` command to delete the harbor operator deployed with helm chart:

```shell
helm uninstall harbor-operator --namespace harbor-operator-ns
```

## Additions

If you selectively decide to install the operators of the dependent services (such as PostgreSQL, Redis and Minio) to achieve the capabilities of deploying full stack Harbor (harbor components + in-cluster dependent services) with helm charts,
you can check the additional references listed below.

* [Install Minio operator with chart](https://github.com/minio/operator/tree/master/helm/operator)
  * Find archived minio operator charts from [here](https://github.com/minio/operator/tree/master/helm-releases)
* [Install PostgreSQL operator with chart](https://github.com/zalando/postgres-operator/blob/master/docs/quickstart.md#helm-chart)
* [Install Redis operator with chart](https://github.com/spotahome/redis-operator#using-the-helm-chart)

Besides, you can also enable the operators of the dependent services in `charts/harbor-operator/values.yaml` to deploy full stack Harbor more efficient.

* Find the configuration items of the dependent operators charts from [here](https://github.com/goharbor/harbor-operator/blob/master/charts/harbor-operator/values.yaml#L252)

## What's next

If the Harbor operator is successfully installed, you can install harbor sample

```shell
kustomize build --reorder legacy 'config/samples/harborcluster' | kubectl apply -f -
```

or use

```shell
kubectl create ns cluster-sample-ns
kubectl config set-context --current --namespace=cluster-sample-ns
make sample-harborcluster
```

or follow the guide shown [here](../tutorial.md#deploy-harbor-cluster) to deploy your Harbor cluster to your Kubernetes cluster and try it.
