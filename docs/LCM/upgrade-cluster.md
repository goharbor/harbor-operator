# How to upgrade your Harbor cluster

The harbor operator only serves a specific version of the harbor and allows to upgrade the harbor cluster instances from the old version to the latest version the operator served.

Assume that the harbor operator v1.0.0 which serves harbor v2.2.1 is installed in the Kubernetes cluster, and there is harbor clusters deployed in the Kubernetes cluster. Now the harbor operator v1.0.1 which serves harbor v2.2.2 is released, and the users want to upgrade the three harbor clusters to v2.2.2. Here are the steps to upgrade the harbor cluster managed by the harbor operator to the new version.

1. Upgrade the harbor operator to v1.0.1 by helm or kustomize, the method depends on the original way to install the harbor operator. [Installation](../installation/installation.md) is a reference resources to upgrade the harbor operator to new release.

2. Edit the manifest of the harbor cluster by `kubectl` and set the `version` field to `2.2.2` in the spec.

   ```bash
   kubectl -n harbor-cluster-ns edit harborclusters cluster-name
   ```

3. The harbor operator will get an update event of the harbor cluster resource and reconcile to upgrade the harbor cluster to v2.2.2.
