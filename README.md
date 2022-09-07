# Harbor Operator

[![CI Pipeline](https://github.com/goharbor/harbor-operator/actions/workflows/tests.yml/badge.svg)](https://github.com/goharbor/harbor-operator/actions/workflows/tests.yml)
[![CodeQL](https://github.com/goharbor/harbor-operator/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/goharbor/harbor-operator/actions/workflows/codeql-analysis.yml)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/bb3adb454b424e66ae3b9bdf2ab2fce1)](https://www.codacy.com/gh/goharbor/harbor-operator/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=goharbor/harbor-operator&amp;utm_campaign=Badge_Grade)
[![Go Reference](https://pkg.go.dev/badge/github.com/goharbor/harbor-operator.svg)](https://pkg.go.dev/github.com/goharbor/harbor-operator)

> **ATTENTIONS:** THE `MASTER` BRANCH MAY BE IN AN UNSTABLE OR EVEN BROKEN STATE DURING DEVELOPMENT.

[Harbor](https://github.com/goharbor/harbor/) is a CNCF hosted open source trusted cloud-native registry project that stores, signs, and scans content. Harbor is composed on numerous stateful and stateless components and dependencies that may be deployed, updated, healed, backuped or scaled respecting some constraints.

The Harbor Operator provides an easy and solid solution to deploy and manage a full Harbor service stack including both the harbor service components and its relevant dependent services such as database, cache and storage services to the target [Kubernetes](https://kubernetes.io/) clusters in a scalable and high-available way. The Harbor Operator defines a set of Harbor-related custom resources on top of Kubernetes [Custom Resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/). The Kubernetes API can then be used in a declarative way to manage Harbor deployment stack and ensure its scalability and high-availability operation, thanks to the [Kubernetes control loop](https://kubernetes.io/docs/concepts/#kubernetes-control-plane). The project harbor-operator aims to cover both `Day1` and `Day2` operations of an enterprise-grade Harbor deployment.

## Features

Harbor deployment stack is controlled by a custom Harbor resource `HarborCluster`. HarborCluster owns the custom resource `Harbor` that represents the Harbor own service stack, and the custom resources of the related dependent services (PostgreSQL, Redis and MinIO) that are required when deploying the full Harbor deployment stack.

* Provides strong flexibility to deploy different stacks of Harbor cluster (identified by `HarborCluster` CR)
  * **Minimal stack:** only required Harbor components `Core`, `Registry`, `Registry Controller`, `Job Service` and `Portal` are provisioned.
  * **Standard stack:** the optional Harbor components `Notary`, `Trivy`, `ChartMuseum` and `Metrics Exporter` can be selected to enable.
  * **Full stack:** both the Harbor components (required+optional) and also the related dependent services including the database (PostgreSQL), cache (Redis) and storage (MinIO) can be deployed into the Kubernetes cluster together with a scalable and high-available way.
* Supports configuring either external or in-cluster deployed dependent services
* Supports a variety of backend storage configurations
  * [X] filesystem: A storage driver configured to use a directory tree in the a kubernetes volume.
  * [X] [s3](https://docs.aws.amazon.com/AmazonS3/latest/API/Welcome.html): A driver storing objects in an Amazon Simple Storage Service (S3) bucket.
  * [X] [swift](https://wiki.openstack.org/wiki/Swift): A driver storing objects in Openstack Swift.
  * [X] [azure](https://azure.microsoft.com/services/storage/): A driver storing objects in Microsoft Azure Blob Storage.
  * [X] [gcs](https://cloud.google.com/storage): A driver storing objects in a Google Cloud Storage bucket.
* Supports updating the deployed Harbor cluster
  * Adjust replicas of components
  * Add/remove the optional Harbor components
* Supports upgrading the managed Harbor registry version
* Deletes all the linked resources when deleting the Harbor cluster
* Support services exposed with [ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/): [nginx(default)](https://git.k8s.io/ingress-nginx/README.md#readme), [gce](https://git.k8s.io/ingress-gce/README.md#readme), [contour](https://github.com/projectcontour/contour) and `ncp`
* Support Day2 operations
  * Configures Harbor system settings with configuration CRD (`recommend`) or labeled ConfigMap (`deprecated`)

## Future features

* Support Day2 operations
  * Image pulling secret auto-injection
    * Auto mapping Kubernetes namespaces to the Harbor project
  * Image pulling path auto-rewriting
    * Transparent proxy cache settings
  * Certification auto injection
  * Manage Harbor resources with the declaration way
    * Robot account
    * Project
    * and more
* [Auto-scaling](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) for each component.
* Backup/restore data (registry layer, chartmuseum data, databases content).
* Support services exposed with LoadBalancer

## Release plans

* [Release 1.3](https://github.com/goharbor/harbor-operator/projects/9)
* [Release 1.4](https://github.com/goharbor/harbor-operator/projects/12)

## Getting started

For a quick first try follow the instructions of this [tutorial](./docs/tutorial.md).

## Versioning

Versions of the underlying components are listed below:

|  Components   | Harbor      | MinIO operator | PostgreSQL operator | Redis operator |
|---------------|-------------|----------------|---------------------|----------------|
|  Versions     | 2.5.x `[1]` | 4.4.28         | 1.6.3               | 1.1.1          |

NOTES:

[1] `.x` means all the patch releases of Harbor can be naturally supported in one operator version.

## Compatibility

### Applicative Kubernetes versions

Harbor operator supports two extra Kubernetes versions besides the current latest version (`n-2` pattern):

|    Versions   |        1.21        |        1.22        |        1.23        |         1.24        |
|---------------|--------------------|--------------------|--------------------|---------------------|
| Compatibility | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: |  :heavy_check_mark: |

### Cert manager versions

Harbor operator relies on cert manager to manage kinds of certificates used by Harbor cluster components. Table shown below lists the compatibilities of cert manager versions:

|    Versions   |         1.6[.3]    |           1.7[.3]    |           1.8[.2]    |           1.9[.1]    |
|---------------|--------------------|----------------------|----------------------|----------------------|
| Compatibility | :heavy_check_mark: |  :heavy_check_mark:  |  :heavy_check_mark:  |  :heavy_check_mark:  |

### Ingress controller types

Harbor operator exposes the frontend service with ingress (CRD version: `v1beta1`). Table shown below lists the ingress controller types supported.

|  Ingress Controller  |  Compatibility      |                           Description                      |
|----------------------|---------------------|------------------------------------------------------------|
|  default             | :heavy_check_mark:  | Default ingress controller like NGINX                      |
|  gce                 | :heavy_check_mark:  | Google Cloud Engine ingress controller                     |
|  ncp                 | :heavy_check_mark:  | NSX-T Container plugin ingress controller                  |
|  contour             | :heavy_check_mark:  | Ingress controller that works by deploying the Envoy proxy |

**NOTES:**

  :heavy_check_mark: : support
  :heavy_multiplication_x: : not support
  :o: : not verified (probably support)

## Documentation

* [How it works](./docs/arch/arch.md)
* [Installation](./docs/installation/installation.md)
* [CRD references](./docs/CRD/custom-resource-definition.md)
* [Manifests references](./docs/manifests-reference.md)
* [Customize storage, database and cache services](./docs/installation/customize-storage-db-redis.md)
* [Customize images](./docs/customize-images.md)
* [Day2 configurations](docs/day2/day2-configurations.md)
* [Upgrade Harbor cluster](./docs/LCM/upgrade-cluster.md)
* [Delete Harbor cluster](./docs/LCM/cluster-deletion.md)
* [Backup data](./docs/LCM/backup-data.md)
* [Useful Makefile Targets](./docs/makefile.md)
* [Operator configurations](./docs/configurations/operator-configurations.md)
* [Enable Minio console](./docs/installation/enable_minio_console.md)
* [Performance comparison between fs & MinIO](./docs/perf/simple-perf-comprasion.md)
* [Monitor logs with grafana](./docs/LCM/monitor-logs-with-grafana.md)

## Contributions

Harbor operator project is developed and maintained by the [Harbor operator workgroup](https://github.com/goharbor/community/blob/master/workgroups/wg-operator/README.md). If you're willing to join the group and do contributions to operator project, welcome to [contact us](#community). Follow the [Development guide](https://github.com/goharbor/harbor-operator/blob/master/docs/development.md) to start on the project.

Special thanks to the [contributors](./MAINTAINERS) who did significant contributions ([see feature area](./docs/feature-areas.md)).

## Community

* **Slack:** channel `#harbor-operator-dev` at [CNCF Workspace](https://slack.cncf.io)
* **Mail group:** send mail to Harbor dev mail group: harbor-dev@lists.cncf.io
* **Twitter:** [@project_harbor](https://twitter.com/project_harbor)
* **Community meeting:** attend [bi-weekly community meeting](https://github.com/goharbor/community/blob/master/MEETING_SCHEDULE.md) for Q&A

## Additional references

* [cert-manager](https://cert-manager.io/docs/)
* [Underlying postgreSQL operator](https://github.com/zalando/postgres-operator)
* [Underlying redis operator](https://github.com/spotahome/redis-operator)
* [Underlying minio operator](https://github.com/minio/minio-operator)
* [Kubernetes operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
* [Custom Resource Definition](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
* [Kubebuilder](https://book.kubebuilder.io/)

## Related links

* Contribute: <https://github.com/goharbor/harbor-operator/blob/master/CONTRIBUTING>
* Report bugs: <https://github.com/goharbor/harbor-operator/issues>
* Get latest version: <https://hub.docker.com/r/goharbor/harbor-operator>

## Recognition

The operator was initially developed by [OVHcloud](https://ovhcloud.com) and donated to the [Harbor](https://github.com/goharbor) community as one of its governing projects in March 2020, becoming the basis of the official Harbor [Operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

OVHcloud uses the operator at scale to operate part of its private registry service. But the operator was designed in an agnostic way and it's continuing to evolve into a more pervasive architecture to bring values to any companies in search of deploying and managing one or multiple Harbor.

## License

See [LICENSE](https://github.com/goharbor/harbor-operator/blob/master/LICENSE) for licensing details.
