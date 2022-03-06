# How to upgrade your Harbor cluster

A specified version of Harbor operator supports a corresponding Harbor minor version as well as its subsequent patch releases. e.g: harbor operator v1.0.1 supports 2.2.x harbor versions such as 2.2.0, 2.2.1 and 2.2.2 etc. For upgrading your Harbor cluster, there might be two different cases.

## Upgrade to patch releases

The guideline shown below describes how to upgrade your Harbor cluster from lower patch version to higher patch version without operator upgrading (because there is no need as a Harbor operator version supports all patch releases).

Assume that the harbor operator v1.1.1 which serves harbor v2.3.x is installed in the Kubernetes cluster, and there is a harbor cluster v2.3.2 deployed in the Kubernetes cluster.

If you want to upgrade the harbor cluster from v2.3.2 to v2.3.5, just edit the manifest of the harbor cluster by `kubectl` and set the `version` field from `2.3.2` to `2.3.5` and the harbor operator will upgrade the harbor cluster instance to harbor v2.3.5.

## Upgrade to minor+ releases

For upgrading Harbor cluster across different minor versions, an operator upgrading should be involved first (because one Harbor operator version only serves one minor version serials). Steps shown below describes how to do such upgrading.

1. Upgrade the harbor operator to the newer version that supports the Harbor version you're planning to upgrade your existing Harbor cluster to by `helm` or `kustomize`, the method depends on the original way to install the harbor operator. [Installation](../installation/installation.md) is a reference resources to upgrade the harbor operator to new release.

1. Edit the manifest of the harbor cluster by `kubectl` and set the `version` field to the newer Harbor version (e.g:`2.3.5`) in the spec.

   ```bash
   kubectl -n harbor-cluster-ns edit harborclusters cluster-name
   ```

1. The harbor operator will get an update event of the harbor cluster resource and reconcile to upgrade the harbor cluster to v2.3.0.
