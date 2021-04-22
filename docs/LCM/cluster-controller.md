# Harbor cluster controller

Cluster operator is responsible for controlling the reconciling process of CR `HarborCluster` that is a top level resource owning the `Harbor` CR and the CRs of Harbor's dependent services. Its primary flow is apply the `Harbor` CR and start `Harbor` CR's reconciling process when the related dependent services CR required in the spec are ready. The apply processes of the related dependent services CR are handled in a concurrent way.

## Resource status

Combination of the conditions that indicates the status of `Harbor` as well as status of its dependent service CRs such as `postgresql`, `redisfailover` and minio `tenant` leads to the current state of the `HarborCluster` resource.

### Conditions

Here are the conditions involved in the status combination process:

|    Condition   |  Description  |
|----------------|---------------|
| `StorageReady` | in-cluster storage minio service is ready |
| `DatabaseReady` | in-cluster database postgresql is ready |
| `CacheReady` | in-cluster cache redis is ready |
| `ServiceReady` | the harbor is ready |

> NOTE: condition `ConfigurationReady` indicates if the cfgMap based Day2 configuration has been successfully applied. But it does not take account into the combination of the overall status.

### Overall status

The overall status dimensions and the conditions of its establishment is listed in the table shown below:

|     Status  | Conditions |
|-------------|------------|
| `healthy`   | 4 conditions are all `True` |
| `unhealthy` | at least one of the 4 conditions is `False` |
| `creating`  | Otherwise |

## Reconcile loop

The reconcile loop is triggered by a Kubernetes event when the `HarborCluster` Resource or relevant resources owned by that `HarborCluster` changes.

```text
                  Event received      +----------------------+
                  ------------------->|      Reconcile       |
                                      +----------+-----------+
                                                 |
                                                 |          True
                                               Deleted? ------------>Exit
                                                 |
                                                 |
                                                 v
                                     Start service deploy group
                 +-------------------------------+----------------------------+
                 |                               |                            |
                 |                               |                            |
                 |                               |                            |
+----------------v---------------+   +-----------+-----------+    +-----------v---------+
|     PostgreSQL Reconcile       |   |    Redis Reconcile    |    |   Minio Reconcile   |
+----------------+---------------+   +-----------+-----------+    +-----------+---------+
                 |                               |                            |
                 |                               |                            |
                 |                               |                            |
                 |                               |                            |
                 v                               v                            v
          Apply or update                  Apply or update             Apply or update
                 |                               |                            |
                 |                               |                            |
                 |                               |                            |
                 |                               |                            |
                 |                               v                            |
                 +--------------------> Any error occurred?<------------------+
                                                 |      |
                                                 | NO   |             YES
                                                 |      +----------------------------> Exit
                                                 |                                      |
                                                 v                   NO                 |
                                      All services are ready?------------------> Exit   |
                                                 |                                |     |
                                                 | YES                            |     |
                                                 |                                |     |
                                                 |                                |     |
                                                 v                                |     |
                                      +----------------------+                    |     |
                                      |   Harbor Reconcile   |                    |     |
                                      +----------+-----------+                    |     |
                                                 |                                |     |
                                                 |                                |     |
                                                 |                                |     |
                                                 v                                |     |
                                          Apply or update?                        |     |
                                                 |                                |     |
                                                 |                                |     |
                                                 |                                |     |
                                                 |                                |     |
                                                 v              YES               |     |
                                        Any error occurred? -----------> Exit     |     |
                                                 |                        |       |     |
                                                 |                        |       |     |
                                                 | NO                     |       |     |
                                                 |                        |       |     |
                                                 v                        |       |     |
                                             Completed                    |       |     |
                                                 |                        |       |     |
                                                 |                        |       |     |
                                                 |                        |       |     |
                                                 |                        |       |     |
                                  +--------------+--------------+         |       |     |
                                  |  Set status & conditions    <---------+-------+-----+
                                  +-----------------------------+
```
