# Tutorial

## Install Harbor operator stack

TBD

## Deploy Harbor cluster

TBD

## Try the deployed Harbor cluster

TBD

## Deploy the sample

1. Deploy the Harbor resource with `make sample`.  
   But do not hesitate to edit the resource once deployed `kubectl edit harbor harbor-sample`.

   Then check that Harbor is deployed. Note: Plugins such as [kubectl-tree](https://github.com/ahmetb/kubectl-tree) are nice to have a better overview.

   ```bash
   kubectl get po
   ```

2. Get the certificate authority used to generate the public certificate and install it on your computer (on the system scope, docker daemon + browser).

   ```bash
   kubectl get secret public-certificate -o jsonpath='{.data.ca\.crt}' \
     | base64 --decode
   ```

3. Access to Portal with the publicURL `kubectl get harbor sample -o jsonpath='{.spec.externalURL}''.
   Connect with the admin user and with the following password.

   ```bash
   kubectl get secret "$(kubectl get harbor sample -o jsonpath='{.spec.harborAdminPasswordRef}')" -o jsonpath='{.data.secret}' \
     | base64 --decode
   ```

Few customizations are available:

- [Custom Registry storage](samples/registry-storage-configuration.md)
- [Database configuration](samples/database-installation.md)
- [Redis configuration](samples/redis-installation.md)