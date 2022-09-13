# harbor-operator

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.0.0-dev](https://img.shields.io/badge/AppVersion-0.0.0--dev-informational?style=flat-square)

Deploy Harbor Operator

**Homepage:** <https://github.com/goharbor/harbor-operator>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Expects input structure as per specification <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#affinity-v1-core> For example: `{   "nodeAffinity": {     "requiredDuringSchedulingIgnoredDuringExecution": {       "nodeSelectorTerms": [         {           "matchExpressions": [             {               "key": "foo.bar.com/role",               "operator": "In",               "values": [                 "master"               ]             }           ]         }       ]     }   } }` |
| allowPrivilegeEscalation | bool | `false` | Allow privilege escalation for the controller Pods |
| autoscaling.enabled | bool | `false` | Whether to enabled [Horizontal Pod Autoscaling](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) |
| autoscaling.maxReplicas | int | `100` | Maximum conroller replicas |
| autoscaling.minReplicas | int | `1` | Minimum conroller replicas |
| autoscaling.targetCPUUtilizationPercentage | int | `80` | CPU usage target for autoscaling |
| autoscaling.targetMemoryUtilizationPercentage | int | No target | Memory usage target for autoscaling |
| controllers.chartmuseum.maxReconcile | int | `1` | Max parallel reconciliation for ChartMuseum controller |
| controllers.common.classname | string | `""` | Harbor class handled by the operator. An empty class means watch all resources |
| controllers.common.networkPolicies | bool | `false` | Whether the operator should manage network policies |
| controllers.common.watchChildren | bool | `true` | Whether the operator should watch children |
| controllers.core.maxReconcile | int | `1` | Max parallel reconciliation for Core controller |
| controllers.harbor.maxReconcile | int | `1` | Max parallel reconciliation for Harbor controller |
| controllers.harborConfiguration.maxReconcile | int | `1` | Max parallel reconciliation for HarborConfiguration controller |
| controllers.harborcluster.maxReconcile | int | `1` | Max parallel reconciliation for HarborCluster controller |
| controllers.jobservice.maxReconcile | int | `1` | Max parallel reconciliation for JobService controller |
| controllers.notaryserver.maxReconcile | int | `1` | Max parallel reconciliation for NotaryServer controller |
| controllers.notarysigner.maxReconcile | int | `1` | Max parallel reconciliation for NotarySigner controller |
| controllers.portal.maxReconcile | int | `1` | Max parallel reconciliation for Portal controller |
| controllers.registry.maxReconcile | int | `1` | Max parallel reconciliation for Registry controller |
| controllers.registryctl.maxReconcile | int | `1` | Max parallel reconciliation for RegistryCtl controller |
| controllers.trivy.maxReconcile | int | `1` | Max parallel reconciliation for Trivy controller |
| deploymentAnnotations | object | `{}` | Additional annotations to add to the controller Deployment |
| fullnameOverride | string | `""` |  |
| harborClass | string | `""` | Class name of the Harbor operator |
| image.pullPolicy | string | `"IfNotPresent"` | The image pull policy for the controller. |
| image.repository | string | `"goharbor/harbor-operator"` | The image repository whose default is the chart appVersion. |
| image.tag | string | `"dev_master"` | The image tag whose default is the chart appVersion. |
| imagePullSecrets | list | `[]` | Reference to one or more secrets to be used when pulling images <https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/> For example: `[   {"name":"image-pull-secret"} ]` |
| installCRDs | bool | `false` | If true, CRD resources will be installed as part of the Helm chart. If enabled, when uninstalling CRD resources will be deleted causing all installed custom resources to be DELETED |
| leaderElection.namespace | string | `"kube-system"` | The namespace used to store the ConfigMap for leader election |
| logLevel | int | `4` | Set the verbosity of controller. Range of 0 - 6 with 6 being the most verbose. Info level is 4. |
| minio-operator.enabled | bool | `false` | Whether to enabled [MinIO Operator](https://github.com/minio/operator) |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` | Expects input structure as per specification <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#nodeselector-v1-core> For example: `[   {     "matchExpressions": [       {         "key": "kubernetes.io/e2e-az-name",         "operator": "In",         "values": [           "e2e-az1",           "e2e-az2"         ]       }     ]   } ]` |
| podAnnotations | object | `{}` | Additional annotations to add to the controller Pods |
| podLabels | object | `{}` | Additional labels to add to the controller Pods |
| podSecurityContext | object | `{"runAsNonRoot":true,"runAsUser":65532}` | Expects input structure as per specification <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#podsecuritycontext-v1-core> For example: `{   "fsGroup": 2000,   "runAsUser": 1000,   "runAsNonRoot": true }` |
| postgres-operator.configKubernetes.secret_name_template | string | `"{username}.{cluster}.credentials"` |  |
| postgres-operator.enabled | bool | `false` | Whether to enabled [Postgres operator](https://github.com/zalando/postgres-operator) |
| priorityClassName | string | `""` | priority class to be used for the harbor-operator pods |
| rbac.create | bool | `true` | Whether to install Role Based Access Control |
| redis-operator.enabled | bool | `false` | Whether to enabled [Redis Operator](https://github.com/spotahome/redis-operator) |
| redis-operator.image.tag | string | `"v1.2.0"` |  |
| replicaCount | int | `1` | Number of replicas for the controller |
| resources | object | `{"limits":{"cpu":"500m","memory":"300Mi"},"requests":{"cpu":"300m","memory":"200Mi"}}` | Expects input structure as per specification <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#resourcerequirements-v1-core> |
| service.port | int | `443` | Expose port for WebHook controller |
| service.type | string | `"ClusterIP"` | Service type to use |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| strategy | object | `{}` | Expects input structure as per specification <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#deploymentstrategy-v1-apps> For example: `{   "type": "RollingUpdate",   "rollingUpdate": {     "maxSurge": 0,     "maxUnavailable": 1   } }` |
| tolerations | list | `[]` | Expects input structure as per specification <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#toleration-v1-core> For example: `[   {     "key": "foo.bar.com/role",     "operator": "Equal",     "value": "master",     "effect": "NoSchedule"   } ]` |
| volumeMounts | list | `[]` | Expects input structure as per specification <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volumemount-v1-core> For example: `[   {     "mountPath": "/test-ebs",     "name": "test-volume"   } ]` |
| volumes | list | `[]` | Expects input structure as per specification <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volume-v1-core> For example: `[   {     "name": "test-volume",     "awsElasticBlockStore": {       "volumeID": "<volume-id>",       "fsType": "ext4"     }   } ]` |
