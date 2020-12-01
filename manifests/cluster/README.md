# Deployment manifests

Manifest yaml files here are `kustomization` templates used to deploy the harbor operator as well its dependant operators.

## Notes

* Deployments of `Cert Manager` and `Ingress Controller` are not covered yet.
* Other relying on operators are linked via remote manifests.
* `deployment.yaml` is the all-in-one operator deployment manifest yaml file built by the `kustomization` templates.

## Usage

### Deploy

Directly use deployment yaml:

```shell script
kubectl apply -f ./deployment.yaml
```

Use `kustomize`:

```shell script
kustomize build . | kubectl apply -f -
```

### Uninstall

Use manifest:

```shell script
kubectl delete -f ./deployment.yaml
```

Use `kustomize`:

```shell script
kustomize build . | kubectl delete -f -
```
