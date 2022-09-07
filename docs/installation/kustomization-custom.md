# Install with manual steps

> NOTES: Harbor operator relies on other service operators to deploy the in-cluster dependent services (PostgreSql, storage(Minio) and Redis) for the deploying Harbor cluster. As the Harbor cluster also supports configuring the existing pre-installed services as its dependent services, if you can confirm your cluster users has no needs to deploy a full stack Harbor cluster (harbor components + in-cluster PostgreSQL & Redis & Minio), then some service operators can be skipped to deploy. Under this situation, the [all-in-one way](./kustomization-all-in-one.md) will not be applicable to you. You can install the harbor operator stack per your own demands.

The installation guide documented here help you deploy Harbor operator stack with manual steps.

## Prerequisites

Check the list shown [here](./kustomization-all-in-one.md#prerequisites).

## Deploy PostgreSQL operator (Optional)

Follow the installation guide shown [here](https://github.com/zalando/postgres-operator/blob/master/docs/quickstart.md#configuration-options) to install the PostgreSQL operator.It can be used with kubectl 1.14 or newer as easy as:

```shell script
kubectl apply -k github.com/zalando/postgres-operator/manifests
```

Check the PostgreSQL operator status (by default it's deployed into the `default` namespace):

```shell
kubectl get pod -l name=postgres-operator
```

Output:

```log
NAME                                 READY   STATUS    RESTARTS   AGE
postgres-operator-6cc989d674-ddm6n   1/1     Running   0          56s
```

For deleting the PostgreSQL operator, just call:

```shell
kubectl delete -k github.com/zalando/postgres-operator/manifests
```

## Deploy Redis operator (Optional)

Follow the deployment guide shown [here](https://github.com/spotahome/redis-operator/tree/v1.0.1#operator-deployment-on-kubernetes) to deploy Redis operator to your cluster.

A simple way is:

```shell script
kubectl create -f https://raw.githubusercontent.com/spotahome/redis-operator/master/example/operator/all-redis-operator-resources.yaml
```

Check the Redis operator status (by default it's deployed into the `default` namespace):

```shell
kubectl get pod -l app=redisoperator
```

Output:

```log
NAME                            READY   STATUS    RESTARTS   AGE
redisoperator-56d6888cc-5sz9k   1/1     Running   0          84s
```

For deleting the Redis operator, just call:

```shell
kubectl delete -f https://raw.githubusercontent.com/spotahome/redis-operator/master/example/operator/all-redis-operator-resources.yaml
```

## Deploy Minio operator (Optional)

Follow the installation guide shown [here](https://github.com/minio/operator#deploy-the-minio-operator-and-create-a-tenant) to install the Minio operator.

Or use the Minio kustomization template:

```shell
# Clone the codebase.
git clone https://github.com/minio/operator.git

# Apply with the kustomization template that is located the root dir of the codebase.
kustomize build | kubectl apply -f -
```

Check the Minio operator status (by default it's deployed into the `minio-operator` namespace):

```shell
kubectl get pod -n minio-operator
```

Output:

```log
NAME                              READY   STATUS    RESTARTS   AGE
console-6899978d9f-vrwnj          1/1     Running   0          46s
minio-operator-868d7466fc-ppw96   1/1     Running   0          45s
```

For deleting the Minio operator, call

```shell
kustomize build | kubectl delete -f -
```

## Deploy Harbor operator

Deploy the Harbor operator with the deployment manifest:

```shell
kubectl apply -f https://raw.githubusercontent.com/goharbor/harbor-operator/master/manifests/harbor/deployment.yaml
```

Check the Harbor operator status (by default it's deployed into the `harbor-operator-ns` namespace):

```shell
kubectl get pod -n harbor-operator-ns
```

Output:

```shell
NAME                               READY   STATUS    RESTARTS   AGE
harbor-operator-76c44d8ddd-z7rgx   1/1     Running   0          80s
```

For deleting the Harbor operator, call

```shell
kubectl delete -f https://raw.githubusercontent.com/goharbor/harbor-operator/master/manifests/harbor/deployment.yaml
```

Of course, you can clone the codebase into your computer and then customize and deploy with the kustomization template:

```shell
git clone https://github.com/goharbor/harbor-operator.git

# Checkout to the specified branch or the specified tag.
# To branch: git checkout <branch-name> e.g.: git checkout release-1.3.0
# To tag: git checkout tags/<tag> -b <branch-name> e.g: git checkout tags/v1.3.0 -b tag-v1.3.0

# As the resource manifests are not stored in the codebase, then you need to run the following command to generate the related resource manifests before using `kustomize` to build your customized operator deployment manifest:
make manifests

# Use kustomization template to deploy the Harbor operator.
kustomize build manifests/harbor | kubectl apply -f -

# Delete the Harbor operator.
## kustomize build manifests/harbor | kubectl delete -f -
```

>NOTES: Here we use the deployment manifest in the `main` branch as an example, for deploying the released versions,you can get the deployment manifest in the GitHub release page or find it in the corresponding code branch such as `release-1.3.0`.

## What's next

If the Harbor operator is successfully installed, you can follow the guide shown [here](../tutorial.md#deploy-harbor-cluster) to deploy your Harbor cluster to your Kubernetes cluster and try it.
