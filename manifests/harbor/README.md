# Deployment manifests

Manifest yaml files here are `kustomization` templates used to deploy only the harbor operator.

## Notes

* `deployment.yaml` is the operator deployment manifest yaml file built by the `kustomization` templates.

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
