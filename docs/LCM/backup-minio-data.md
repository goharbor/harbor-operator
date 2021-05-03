# Backup Minio data

Documentation shown here guides you how to backup the `Minio` data.

>NOTES: The document shown here is for reference only, you can also look for other available data backup alternatives based on your real case.

## Prerequisites

1. You need to install [mc]( https://docs.min.io/docs/minio-client-complete-guide.html) that is the MinIO CLI.

2. Get the `access_key` and `secret_key` of MinIO.

    ```shell
    $ kubectl get secret minio-access-secret -n cluster-sample-ns -o jsonpath='{.data.accesskey}' | base64 --decode
    admin
    $ kubectl get secret minio-access-secret -n cluster-sample-ns -o jsonpath='{.data.secretkey}' | base64 --decode
    minio123
    ```

3. Please make sure you have the correct MinIO address to access.

## Backup

### Step 1

Create the directory to storage Harbor Data in the backup pod, please make sure you have the Read/Write Permissions.

```shell
mkdir {minio_backup_directory}
```

### Step 2

Setup mc config.

```shell
mc --insecure -C {minio_backup_directory} config host add minio \ http(s)://{minio_address}:9000 {minio_access_key} {minio_secret_key}
```

### Step 3

Start backup minio data

```shell
mc --insecure -C {minio_backup_directory} \  
cp -r minio/harbor {minio_backup_directory}
```

### Step 4

Make sure that all data are backuped in the {minio_backup_directory}.
