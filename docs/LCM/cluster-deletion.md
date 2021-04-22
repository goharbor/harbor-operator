# Delete the Harbor cluster

For deleting the deployed Harbor cluster, one important thing to note here is, you should make sure all your data is correctly [backup](./backup-data.md) before deletion if you cares about your data (not development or testing environments).

## Delete the Harbor cluster CR

You can delete your Harbor cluster by deleting the CR with `kubectl`:

```shell
# Replace the text between <> with the real ones.
kubectl delete HarborCluster/<myClusterName> -n <myNamespace>

# e.g:
# kubectl delete HarborCluster/harbor-cluster-sample -n sample
```

> NOTE: the command shown above will delete the Harbor cluster as well as all the associated resources (including all the PVs) except the ones pre-created such as namespace, root password or cert-manager issuer etc.

## Delete by manifest

The Harbor cluster can also be deleted via the deployment manifests that may be either the manifest yaml file or the kustomization template file. With this way, both the associated resources and the resources pre-defined in the manifest like namespace, root password or cert-manager issuer etc. can be cleaned up at the same time.

```shell
# With deployment manifest
kubectl delete -f my-deployment.yaml

# OR with kustomization template
kustomize build | kubectl delete -f -
```
