# Harbor Reconciler

## Resource status

Each statuses have 3 states: `unknown`, `true`, `false`.
On [Reconciler](#control-loop) side, `unknown` is considered the same way than `false`.

### Applied

Harbor component expose a `applied` status (see it with `kubectl get harbor -o wide`). This status is computed thanks to success/error when applying changes. When updating Harbor resource or [its children](https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#owners-and-dependents) the operator apply changes and update the `applied` status.

### Ready

Harbor component expose a `ready` status (see it with `kubectl get harbor -o wide`). This status is computed thanks to the result of a call to Harbor Core on  `/api/health`.
The component which is not ready is displayed in the status message

```bash
kubectl describe harbor
```

## Control loop

```text
                +--------------+
+-------------> | Control loop |
|               +------+-------+
|                      |
|                   Deleted?      -----> Exit
|                      |           True
|                      v
|               Generation == 0   -----> Patch with default value: conversion
|                      |           True         webhook does not work
|                      v
|    (1)        Check readiness   -----> Update status
|                      |
|                      v
|    (2)        Same Generation?  -----> Applied to false
|                      |          False
|                      v
|        +-------  Applied?  -------+
|        |False                 True|
|        v                          v
|      Apply                     Create
| Applied to True                   |
|        |                          |
|        +-------------+------------+
|                      |
|                      v
|       Ready (1) & Same generation (2)  -----> Exit
|                      |                  True
+----------------------+
```
