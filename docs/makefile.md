# Makefile

There are some useful `make` targets for you to trigger some actions.

## Prerequisites

Clone the harbor operator codebase into your computer with command

```shell
git clone https://github.com/goharbor/harbor-operator.git
```

The `Makefile` is in the root dir of the code repository.

## Useful Makefile targets

|     Target     |      Description      |
|----------------|-----------------------|
| `helm-install` | Install Harbor operator from chart source |
| `helm-generate`| Generate Harbor operator helm chart tgz package |
| `docker-build` | Build operator image from source |
| `docker-push`  | Push the image built by `docker-build` to the repository |
| `install`      | Install CRDs into the cluster |
| `uninstall`    | Uninstall CRDs from the cluster |
| `install-dependencies` | Install the related dependencies including cert-manager, ingress controller, redis and postgresql |
| `dev-tools`    | Install kids of the development tools |
| `sample-%`    | Deploy the related sample CR. `%` can be the name of sub folders under [samples](../config/samples) |
