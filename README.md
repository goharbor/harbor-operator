# Harbor Operator

**ATTENTIONS: THIS PROJECT IS STILL UNDER DEVELOPMENT AND NOT STABLE YET.**

## Why an Harbor Operator

[Harbor](https://github.com/goharbor/harbor/) is a very active project, composed on numerous stateful and stateless sub-projects and dependencies.
These components may be deployed, updated, healed, backuped or scaled respecting some constraints.

The Harbor Operator extends the usual K8s resources with Harbor-related custom ones. The Kubernetes API can then be used in a declarative way to manage Harbor and ensure its high-availability operation, thanks to the [Kubernetes control loop](https://kubernetes.io/docs/concepts/#kubernetes-control-plane).

The Harbor operator aims to cover both Day1 and Day2 operations of an enterprise-grade Harbor deployment.

The operator was initially developed by [OVHcloud](https://ovhcloud.com) and donated to the [CNCF](https://www.cncf.io/) as part of the Harbor project in March 2020, becoming the basis of the official [Kubernetes Operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

## A modular and agnostic design

[OVHcloud](https://ovhcloud.com) uses the operator at scale to operate part of its private registry service, but the project was designed in an agnostic way, to bring value to any company in search of deploying and managing one or multiple Harbor.

Configuration allows tuning both Harbor itself (with or without some optional components) or its dependencies.
It is designed to be used on any Kubernetes cluster, in a cloud or on premise context.

## Project status

Harbor Operator is still very early stage and currently covers deployment, scale and destruction of Harbor in 1.10 version.
Other parts of the life-cycle will be managed in future versions of the operator.
As any project in this repository, do not hesitate to raise issues or suggest code improvements.

## Features

Harbor components is controlled by a custom Harbor resource.
With ConfigMaps and [Secrets](https://kubernetes.io/docs/concepts/configuration/secret/), it handles almost all configuration combination.

### Supported storage configuration

- [X] filesystem: A storage driver configured to use a directory tree in the a kubernetes volume.
- [X] s3: A driver storing objects in an Amazon Simple Storage Service (S3) bucket.
- [X] swift: A driver storing objects in Openstack Swift.
- [ ] azure: A driver storing objects in Microsoft Azure Blob Storage.
- [ ] oss: A driver storing objects in Aliyun OSS.
- [ ] gcs: A driver storing objects in a Google Cloud Storage bucket.

### Deploy a new stack

This operator is able to deploy an Harbor stack, fully or partially.

When Creating the `Harbor` resource, following components are always deployed:

- Harbor Core
- Registry
- Registry Controller
- Portal
- Job Service

Following components are optional:

- ChartMuseum
- Notary
- Clair
- Trivy

### Delete the stack

When deleting the Harbor resource, all linked components are deleted. With two Harbor resources, the right components are deleted and components of the other Harbor are not changed.

### Adding/Removing a component

It is possible to add and delete ChartMuseum, Notary, Clair and Trivy by editing the Harbor resource.

### Future features

1. [Auto-scaling](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) for each component.
2. Backup/restore data (registry layer, chartmuseum data, databases content).

## Installation

See [install documentation](./docs/installation/installation.md).

## Compatibility

### Supported platforms

- [Kubernetes](https://kubernetes.io/docs/concepts/overview/kubernetes-api/) >= 1.16

### Harbor version

This Operator currently only supports Harbor version 2.0

## Howto's

## Configuration

Generate resources using `make generate`

## Development

Now, this project is maintained and developed by the [Harbor operator workgroup](https://github.com/goharbor/community/blob/master/workgroups/wg-operator/README.md). If you're willing to join the group and do contributions to operator project, welcome to [contact us](#community). Follow the [Development guide](https://github.com/goharbor/harbor-operator/blob/master/docs/development.md) to start on the project.

## Community

- Slack channel `#harbor-operator-dev` at [CNCF Workspace](https://slack.cncf.io)
- Send mail to Harbor dev mail group: harbor-dev@lists.cncf.io

## Additional documentation

 1. [Learn how reconciliation works](https://github.com/goharbor/harbor-operator/blob/master/docs/reconciler.md)
 2. [Custom Resource Definition](https://github.com/goharbor/harbor-operator/blob/master/docs/custom-resource-definition.md)

## Related links

- Contribute: <https://github.com/goharbor/harbor-operator/blob/master/CONTRIBUTING.md>
- Report bugs: <https://github.com/goharbor/harbor-operator/issues>
- Get latest version: <https://hub.docker.com/r/goharbor/harbor-operator>

## License

See <https://github.com/goharbor/harbor-operator/blob/master/LICENSE>
