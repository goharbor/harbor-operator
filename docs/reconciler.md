# Harbor Reconciler

## Statuses

Each statuses have 3 states: `unknown`, `true`, `false`.
On Reconciler side, `unknown` is considered the same way than `false`.

- Processing
- Apply
- Ready

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
