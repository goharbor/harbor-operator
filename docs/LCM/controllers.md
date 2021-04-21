# Controllers

ATM All controllers are based on the same framework except [harbor-cluster](./cluster-controller.md).

The framework manages basic status fields: `observedGeneration` and `conditions`.

## Resource status

Combination of observedGeneration and conditions leads to the current state of the object:

### Observed generation

Expose the last generation reconciled by the controller (or currently reconciling).

### Conditions

Kubernetes expose `.status.conditions` which describe the current state of the resource.

Each conditions have 3 states: `unknown`, `true`, `false`.
On [Reconciler](#control-loop) side, `unknown` is considered the same way than `false`.

### 1. In progress

`True` if the controller is currently reconciling the resource.

### Failed

`True` if kubernetes resource is in expected state.

## Control loop

The control loop is triggered by a Kubernetes event when one of the Resource controlled by the operator changes.

```text
Event received  +------------ +
+-------------> |  Reconcile  |
                +------+------+
                       |
     (1)           Deleted? -----------> Exit
                       |            True
                       v
     (2)     Set InProgress = True
        observedGeneration = generation
                       |
                       v
     (3)         Compute all
              resources to apply
                       |
                       v
     (4)       For each resource
                       |
                       v
     (4.1)    Apply if necessary*
                       |
                       v
     (4.1)     Is resource ready? -----> Set Failed = true
                       |           False         |
                       v                         v
     (5)      Set Failed = False           Exit with error
              InProgress = True
                       |
                       v
                      Exit
```

### Note

- Step `(1)`: Nothing to do with deleted resources since all owned resources are handled by kubernetes GC: <https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#owners-and-dependents>
- Step `(2)` and `(3)`: Update `observedGeneration` and
- Step `(3)`: Resources to deploy are computed in memory and added to a graph of dependency.

  - Leaves are resources that should be applied and ready.
  - Roots are resources that *own* resources that should be applied.
  - Branches are leaves and roots: resources that should be applied and resources that own other resources.
  - The main root of the graph is the custom resource.
  - Other roots of the graph are external resources not created by the controller.

- Step `(4.1)`:

  - To check if a resource needs to be applied, the controller compute the checksum of the owner resource (probably the custom resource) and the checksum of the owner when the resource was last applied.
    - The checksum of the last owner is stored in annotations.
    - The checksum is computed thanks to `.metadata.generation` or `.metadata.resourceVersion` of the owner depending of the context.
  - The apply method use [server side apply](https://kubernetes.io/docs/reference/using-api/server-side-apply/).
