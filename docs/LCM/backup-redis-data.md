# Backup Redis data

Documentation shown here guides you how to backup the `Redis` data.

>NOTES: The document shown here is for reference only, you can also look for other available data backup alternatives based on your real case.

## Step 1

Get redis password from harborcluster namespace.

```bash
$ kubectl get secret harborcluster-sample-redis -n cluster-sample-ns -o jsonpath='{.data.password}' | base64 --decode
tfliijxj
```

After above operation, we get the redis password `tfliijxj`.

## Step 2

Execute `SAVE` command and with `-a` specify the password obtained from `step 1`   in redis pod to backup redis data to dump.rdb.

```bash
$ kubectl get po -n cluster-sample-ns |grep redis
rfr-harborcluster-sample-redis-0                                 1/1     Running   0          3h10m
rfs-harborcluster-sample-redis-896dd458d-s6klm                   1/1     Running   0          3h10m

$ kubectl exec rfr-harborcluster-sample-redis-0 -n cluster-sample-ns -- redis-cli SAVE -a tfliijxj
Warning: Using a password with '-a' or '-u' option on the command line interface may not be safe.
OK
```

## Step 3

Copy redis dump.rdb from redis pod to local path.

```bash
$ kubectl cp cluster-sample-ns/rfr-harborcluster-sample-redis-0:dump.rdb dump.rdb
$ ls
dump.rdb
```
