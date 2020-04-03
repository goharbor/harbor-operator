# Redis installation

Many application of the Harbor stack require a [Redis](https://redis.io/).

## With helm

Recommended: Install the [redis helm chart](https://github.com/bitnami/charts/tree/master/bitnami/redis) from bitnami repository.

### Requirements

Checks [chart requirements](https://github.com/bitnami/charts/tree/master/bitnami/redis#prerequisites).

### Step by step

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami

export COMPONENT="chartmuseum"
```

Please repeat following steps for each components requiring a redis: `chartmuseum`, `clair`, `jobservice`, `registry`, ...

1. Install the helm chart

   ```bash
   # Many parameters can be set to configure the redis
   # @see https://github.com/bitnami/charts/tree/master/bitnami/redis#parameters
   helm install "$COMPONENT-redis" bitnami/redis \
     --set usePassword=false
   ```

2. Create the computed secret with correct keys (see [`api/v1alpha1/harbor_secret_format.go`](../../api/v1alpha1/harbor_secret_format.go))

   ```bash
   kubectl create secret "$COMPONENT-redis" \
      --from-literal url="redis://${COMPONENT}-redis-master-0:6379/0" \
      --from-literal namespace=''
   ```

The secret is now ready to use in the *Harbor spec*. Please do previous steps for every components requiring a redis: `chartmuseum`, `clair`, `jobservice`, `registry`, ...

```yaml
apiVersion: goharbor.io/v1alpha1
kind: Harbor
metadata:
  ...
spec:
  ...
  components:
    ...
    registry:
      ...
      cacheSecret: registry-redis
    jobService:
      ...
      redisSecret: jobservice-redis
    clair:
      ...
      adapter:
        ...
        redisSecret: clair-redis
    chartMuseum:
      ...
      cacheSecret: chartmuseum-redis
```
