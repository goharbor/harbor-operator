# Backup Postgresql data

Documentation shown here guides you how to backup the `Postgresql` database data.

>NOTES: The document shown here is for reference only, you can also look for other available data backup alternatives based on your real case.

## get postgresql username and password from secret

- First, we use kubectl search secret which stores the harbor database username and password.

```shell

kubectl get secret -n cluster-sample-ns|grep credentials|grep ^harbor
harbor.postgresql-cluster-sample-ns-ghostbaby-rs.credentials     Opaque                                2      11s

```

- Second, we get the username and password.

```shell

kubectl get secret harbor.postgresql-cluster-sample-ns-ghostbaby-rs.credentials -n cluster-sample-ns -o jsonpath="{.data.username}" | base64 --decode
harbor

kubectl get secret harbor.postgresql-cluster-sample-ns-ghostbaby-rs.credentials -n cluster-sample-ns -o jsonpath="{.data.password}" | base64 --decode
UT6Jo1L8QUhgzB2cqD70xWjwY97fNDxw5vWSULznV0IIinTlEaE9OBypn3UH2uYO

```

## backup postgres harbor database

- find postgres pod

```shell

kubectl get pod -n cluster-sample-ns |grep postgresql
postgresql-cluster-sample-ns-ghostbaby-rs-0                1/1     Running   0          9m19s

```

- login postgres pod

```shell

kubectl exec -it postgresql-cluster-sample-ns-ghostbaby-rs-0 -n cluster-sample-ns -- /bin/bash

```

- save postgres backup shell script

```shell

#!/bin/bash
set -ex
# define postgresql connect url
export PGHOST='127.0.0.1'
export PGPORT='5432'
export PGUSER='harbor'
export PGPASSWORD='UT6Jo1L8QUhgzB2cqD70xWjwY97fNDxw5vWSULznV0IIinTlEaE9OBypn3UH2uYO'
# define backup dir
BASE_DIR='/data/pgsql-dump'
# define backup filename
FILENAME='pgsql-dump'
# define data format，20190329-00
DATE=$(date +%Y%m%d)
TIME=$(date +%H%M)
# render backup full path
BACKUP_DIR="${BASE_DIR}/${DATE}/${TIME}"
mkdir -p ${BACKUP_DIR}
# get backup database list
DATABASES=$(psql --host=${PGHOST} --port=${PGPORT} --username=${PGUSER} -c "\l" | awk '{print$1}' | xargs)
# exec pg_dump to backup data
for DB in $DATABASES;do
    pg_dump --host=${PGHOST} --port=${PGPORT} --username=${PGUSER} --verbose --clean --create ${DB} | gzip -c > ${BACKUP_DIR}/${FILENAME}_${DB}.gz
done
# backup all databases
pg_dumpall --host=${PGHOST} --port=${PGPORT} --username=${PGUSER} --verbose --clean | gzip -c > ${BACKUP_DIR}/${FILENAME}_all.gz
# Find and clean up files whose modification time is greater than 30 days in the backup directory
find ${BASE_DIR}/ -mtime +30 -exec rm -rf {} \;

```

- execute backup script

```shell

bash backup.sh

```

- find backup file

```shell

cd /data/pgsql-dump/

```

- copy database backup file to local

```shell

kubectl cp cluster-sample-ns/postgresql-cluster-sample-ns-ghostbaby-rs-0:/data/pgsql-dump/20210430/1423/pgsql-dump_all.gz ~/pgsql-dump_all.gz

```
