# Installation

## Requirements

Kubernetes API running (see [Supported platforms](https://github.com/goharbor/harbor-core-operator/blob/master/README.md#supported-platforms) for more information).

### Minimal

1. [CertManager](https://docs.cert-manager.io) (version >= 1.11) and an [issuer](https://cert-manager.io/docs/configuration/selfsigned/).
2. Redis for Job Service (such as [Redis HA Helm chart](https://github.com/helm/charts/tree/master/stable/redis-ha)).
3. Core database (such as [PostgreSQL Helm chart](https://github.com/helm/charts/tree/master/stable/postgresql)).
4. Registry storage backend (such as any S3 compatible object storage).

### Additional

1. Ingress controller (such as [nginx Helm chart](https://github.com/helm/charts/tree/master/stable/nginx-ingress)).
2. Clair database (such as [PostgreSQL Helm chart](https://github.com/helm/charts/tree/master/stable/postgresql)).
3. ChartMuseum storage backend (such as any S3 compatible object storage).
4. Notary databases (such as [PostgreSQL Helm chart](https://github.com/helm/charts/tree/master/stable/postgresql)).

## Deploy the operator

1. Get a [kubeconfig file](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) and [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) working.

   ```bash
   kubectl version
   ```

2. Build the application and push it to a registry so it is accessible from Kubernetes nodes.

   ```bash
   make docker-build IMG=the.registry/goharbor/harbor-operator:dev
   make docker-push IMG=the.registry/goharbor/harbor-operator:dev
   ```

3. Deploy requirements

   ```bash
   make install-dependencies
   ```

4. Deploy the application

   ```bash
   make deploy
   ```

5. Ensure the operator is running

   ```bash
   kubectl get po -n harbor-operator-system
   ```

## Deploy the sample

1. Deploy the Harbor resource with `make sample`.  
   But do not hesitate to edit the resource once deployed `kubectl edit harbor harbor-sample`.

   Then check that Harbor is deployed. Note: Plugins such as [kubectl-tree](https://github.com/ahmetb/kubectl-tree) are nice to have a better overview.

   ```bash
   kubectl get po
   ```

2. Get the certificate authority used to generate the public certificate and install it on your computer (on the system scope, docker daemon + browser).

   ```bash
   kubectl get secret public-certificate -o jsonpath='{.data.ca\.crt}' | base64 --decode
   ```

3. Access to Portal with the publicURL `kubectl get harbor harbor-sample -o jsonpath='{.spec.publicURL}'.
   Connect with the admin user and with the following password.

   ```bash
   kubectl get secret "$(kubectl get harbor harbor-sample -o jsonpath='{.spec.adminPasswordSecret}')" -o jsonpath='{.data.password}' | base64 --decode
   ```
