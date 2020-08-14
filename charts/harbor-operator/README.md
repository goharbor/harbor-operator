# harbor-operator

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.0.0-dev](https://img.shields.io/badge/AppVersion-0.0.0-dev-informational?style=flat-square)

Deploy Harbor Operator

**Homepage:** <https://github.com/goharbor/harbor-operator>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Expects input structure as per specification <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#affinity-v1-core> For example: `{   "nodeAffinity": {     "requiredDuringSchedulingIgnoredDuringExecution": {       "nodeSelectorTerms": [         {           "matchExpressions": [             {               "key": "foo.bar.com/role",               "operator": "In",               "values": [                 "master"               ]             }           ]         }       ]     }   } }` |
| autoscaling.enabled | bool | `false` | Whether to enabled [Horizontal Pod Autoscaling](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) |
| autoscaling.maxReplicas | int | `100` | Maximum conroller replicas |
| autoscaling.minReplicas | int | `1` | Minimum conroller replicas |
| autoscaling.targetCPUUtilizationPercentage | int | `80` | CPU usage target for autoscaling |
| autoscaling.targetMemoryUtilizationPercentage | int | No target | Memory usage target for autoscaling |
| certmanager.enabled | bool | `true` | Whether to install cert-manager Helm chart |
| deploymentAnnotations | object | `{}` | Additional annotations to add to the controller Deployment |
| extraEnv | list | `[{"name":"HARBOR_CONTROLLER_MAX_RECONCILE","value":"1"},{"name":"HARBOR_CONTROLLER_WATCH_CHILDREN","value":"true"}]` | Environment variables to inject in controller |
| fullnameOverride | string | `""` |  |
| harborClass | string | `""` | Class name of the Harbor operator |
| image.pullPolicy | string | `"IfNotPresent"` | The image pull policy for the controller. |
| image.repository | string | `"goharbor/harbor-operator"` | The image repository whose default is the chart appVersion. |
| image.tag | string | `"dev"` | The image tag whose default is the chart appVersion. |
| imagePullSecrets | list | `[]` | Reference to one or more secrets to be used when pulling images <https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/> For example: `[   {"name":"image-pull-secret"} ]` |
| ingress.enabled | bool | `false` | Whether to install ingress controller Helm chart |
| leaderElection.namespace | string | `"kube-system"` | The namespace used to store the ConfigMap for leader election |
| logLevel | int | `4` | Set the verbosity of controller. Range of 0 - 6 with 6 being the most verbose. Info level is 4. |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` | Expects input structure as per specification <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#nodeselector-v1-core> For example: `[   {     "matchExpressions": [       {         "key": "kubernetes.io/e2e-az-name",         "operator": "In",         "values": [           "e2e-az1",           "e2e-az2"         ]       }     ]   } ]` |
| podAnnotations | object | `{}` | Additional annotations to add to the controller Pods |
| podLabels | object | `{}` | Additional labels to add to the controller Pods |
| podSecurityContext | object | `{}` | Expects input structure as per specification <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#podsecuritycontext-v1-core> For example: `{   "fsGroup": 2000,   "runAsUser": 1000,   "runAsNonRoot": true }` |
| priorityClassName | string | `""` | priority class to be used for the harbor-operator pods |
| prometheusOperator.enabled | bool | `false` | Whether to install prometheus operator Helm chart |
| rbac.create | bool | `true` | Whether to install Role Based Access Control |
| replicaCount | int | `1` | Number of replicas for the controller |
| resources | object | `{}` | Expects input structure as per specification <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#resourcerequirements-v1-core> `{   "limits": {     "cpu": "100m",     "memory": "128Mi"   },   "requests: {     "cpu": "100m",     "memory": "128Mi"   } }` |
| service.port | int | `443` | Expose port for WebHook controller |
| service.type | string | `"ClusterIP"` | Service type to use |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| strategy | object | `{}` | Expects input structure as per specification <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#deploymentstrategy-v1-apps> For example: `{   "type": "RollingUpdate",   "rollingUpdate": {     "maxSurge": 0,     "maxUnavailable": 1   } }` |
| tolerations | list | `[]` | Expects input structure as per specification <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#toleration-v1-core> For example: `[   {     "key": "foo.bar.com/role",     "operator": "Equal",     "value": "master",     "effect": "NoSchedule"   } ]` |
| volumeMounts | list | `[]` | Expects input structure as per specification <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volumemount-v1-core> For example: `[   {     "mountPath": "/test-ebs",     "name": "test-volume"   } ]` |
| volumes | list | `[]` | Expects input structure as per specification <https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volume-v1-core> For example: `[   {     "name": "test-volume",     "awsElasticBlockStore": {       "volumeID": "<volume-id>",       "fsType": "ext4"     }   } ]` |
