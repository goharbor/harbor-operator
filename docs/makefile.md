# Makefile

There are some useful `make` targets for you to trigger some actions.

## Prerequisites

Clone the harbor operator codebase into your computer with command

```shell
git clone https://github.com/goharbor/harbor-operator.git

# Checkout to the specified branch or the specified tag.
# To branch: git checkout <branch-name> e.g.: git checkout release-1.3.0
# To tag: git checkout tags/<tag> -b <branch-name> e.g: git checkout tags/v1.3.0 -b tag-v1.3.0
```

The `Makefile` is in the root dir of the code repository.

## Useful Makefile targets

|     Target     |      Description      |
|----------------|-----------------------|
| `helm-generate`| Generate Harbor operator helm chart template files |
| `helm-install` | Install Harbor operator from chart source |
| `docker-build` | Build operator image from source |
| `docker-push`  | Push the image built by `docker-build` to the repository |
| `install`      | Install CRDs into the cluster |
| `uninstall`    | Uninstall CRDs from the cluster |
| `install-dependencies` | Install the related dependencies including cert-manager, ingress controller, redis and postgresql |
| `dev-tools`    | Install kids of the development tools |
| `sample-%`     | Deploy the related sample CR. `%` can be the name of sub folders under [config/samples/](../config/samples) |
| `postgresql`   | Deploy a PostgreSQL database with bitnami chart |
| `redis`        | Deploy a Redis database with bitnami chart |
| `sample-github-secret` | Create a secret wrapping the GitHub token read from the env variable `GITHUB_TOKEN` |
