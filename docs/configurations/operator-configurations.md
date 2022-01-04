# Operator configurations

There are some settings for you to configure your Harbor operator deployment.

## common config

- For installing by `kustomize` build, change file `config/config/config.yaml`
- For installing by `helm` chart, change `values.yaml` or change configure map in `charts/harbor-operator/templates/configmap.yaml`

### config.yaml

| key            | description           | default |
|----------------|-----------------------|---------|
| controllers-config-directory | the directory of controllers config files. | /etc/harbor-operator |
| classname | Harbor class handled by the operator. | "" |
| network-policies | Whether the operator should manage network policies. | false |
| watch-children | Whether the operator should watch children. | false |
| jaeger | jaeger configure | "" |
| operator | harbor operator pod configure. include Webhook.Port, Metrics.Address, Probe.Address, LeaderElection.Enabled, LeaderElection.Namespace, LeaderElection.ID | |

## controller config files

- For installing by `kustomize` build, change file `config/config/*-ctrl.yaml`
- For installing by `helm` chart, change `values.yaml` or change configure map in `charts/harbor-operator/templates/configmap.yaml`

### chartmuseum-ctrl.yaml

| key            | description           |
|----------------|-----------------------|
| max-reconcile | max parallel reconciliation. |

### core-ctrl.yaml

| key            | description           |
|----------------|-----------------------|
| max-reconcile | max parallel reconciliation. |

### exporter-ctrl.yaml

| key            | description           |
|----------------|-----------------------|
| max-reconcile | max parallel reconciliation. |

### harbor-ctrl.yaml

| key            | description           |
|----------------|-----------------------|
| max-reconcile | max parallel reconciliation. |

### harborcluster-ctrl.yaml

| key            | description           |
|----------------|-----------------------|
| max-reconcile | max parallel reconciliation. |

### harborconfiguration-ctrl.yaml

| key            | description           |
|----------------|-----------------------|
| max-reconcile | max parallel reconciliation. |

### jobservice-ctrl.yaml

| key            | description           |
|----------------|-----------------------|
| max-reconcile | max parallel reconciliation. |

### notaryserver-ctrl.yaml

| key            | description           |
|----------------|-----------------------|
| max-reconcile | max parallel reconciliation. |

### notarysigner-ctrl.yaml

| key            | description           |
|----------------|-----------------------|
| max-reconcile | max parallel reconciliation. |

### portal-ctrl.yaml

| key            | description           |
|----------------|-----------------------|
| max-reconcile | max parallel reconciliation. |

### registry-ctrl.yaml

| key            | description           |
|----------------|-----------------------|
| max-reconcile | max parallel reconciliation. |

### registryctl-ctrl.yaml

| key            | description           |
|----------------|-----------------------|
| max-reconcile | max parallel reconciliation. |

### trivy-ctrl.yaml

| key            | description           |
|----------------|-----------------------|
| max-reconcile | max parallel reconciliation. |
