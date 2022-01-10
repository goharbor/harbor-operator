# Monitor logs with grafana

Quickly setup the grafana+loki+fluentbit for harbor logs in KIND k8s

- change the host `grafana.harbor.domain` and grafana password `Admin123` below

```bash
kubectl apply -f - <<EOF
---
apiVersion: v1
kind: Namespace
metadata:
  name: logging
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: loki-config
  namespace: logging
  labels:
    k8s-app: loki
data:
  # Configuration files: server, input, filters and output
  # ======================================================
  loki-config.yaml: |
    auth_enabled: false

    server:
      http_listen_port: 3100

    ingester:
      lifecycler:
        address: 127.0.0.1
        ring:
          kvstore:
            store: inmemory
          replication_factor: 1
        final_sleep: 0s
      chunk_idle_period: 1h       # Any chunk not receiving new logs in this time will be flushed
      max_chunk_age: 1h           # All chunks will be flushed when they hit this age, default is 1h
      chunk_target_size: 1048576  # Loki will attempt to build chunks up to 1.5MB, flushing first if chunk_idle_period or max_chunk_age is reached first
      chunk_retain_period: 30s    # Must be greater than index read cache TTL if using an index cache (Default index read cache TTL is 5m)
      max_transfer_retries: 0     # Chunk transfers disabled

    schema_config:
      configs:
        - from: 2020-10-24
          store: boltdb-shipper
          object_store: filesystem
          schema: v11
          index:
            prefix: index_
            period: 24h

    storage_config:
      boltdb_shipper:
        active_index_directory: /tmp/loki/boltdb-shipper-active
        cache_location: /tmp/loki/boltdb-shipper-cache
        cache_ttl: 24h         # Can be increased for faster performance over longer query periods, uses more disk space
        shared_store: filesystem
      filesystem:
        directory: /tmp/loki/chunks

    compactor:
      working_directory: /tmp/loki/boltdb-shipper-compactor
      shared_store: filesystem

    limits_config:
      reject_old_samples: true
      reject_old_samples_max_age: 168h

    chunk_store_config:
      max_look_back_period: 0s

    table_manager:
      retention_deletes_enabled: false
      retention_period: 0s

    ruler:
      storage:
        type: local
        local:
          directory: /tmp/loki/rules
      rule_path: /tmp/loki/rules-temp
      alertmanager_url: http://localhost:9093
      ring:
        kvstore:
          store: inmemory
      enable_api: true
---
apiVersion: v1
kind: Pod
metadata:
    name: loki
    namespace: logging
    labels:
      k8s-app: loki
spec:
  containers:
  - name: loki
    image: ap.cicd.harbor.vmwarecna.net/proxy/grafana/loki:2.1.0
    imagePullPolicy: IfNotPresent
    command:
    - /usr/bin/loki
    - "-config.file=/loki-config/loki-config.yaml"
    ports:
    - containerPort: 3100
      name: loki
      protocol: TCP
    volumeMounts:
    - name: loki-config
      mountPath: /loki-config/
  volumes:
  - name: loki-config
    configMap:
      name: loki-config
---
apiVersion: v1
kind: Service
metadata:
  name: loki
  namespace: logging
spec:
  ports:
  - name: loki
    port: 3100
    protocol: TCP
    targetPort: loki
  selector:
    k8s-app: loki
---
apiVersion: v1
kind: Pod
metadata:
    name: grafana
    namespace: logging
    labels:
      k8s-app: grafana
spec:
  containers:
  - name: grafana
    image: ap.cicd.harbor.vmwarecna.net/proxy/grafana/grafana
    imagePullPolicy: IfNotPresent
    env:
    - name: GF_SECURITY_ADMIN_PASSWORD
      value: Admin123
    ports:
    - containerPort: 3000
      name: grafana
      protocol: TCP
---
apiVersion: v1
kind: Service
metadata:
  name: grafana
  namespace: logging
spec:
  ports:
  - name: grafana
    port: 3000
    protocol: TCP
    targetPort: grafana
  selector:
    k8s-app: grafana
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: fluent-bit
  namespace: logging
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: fluent-bit-read
rules:
- apiGroups: [""]
  resources:
  - namespaces
  - pods
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: fluent-bit-read
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: fluent-bit-read
subjects:
- kind: ServiceAccount
  name: fluent-bit
  namespace: logging
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluent-bit-config
  namespace: logging
  labels:
    k8s-app: fluent-bit
data:
  # Configuration files: server, input, filters and output
  # ======================================================
  fluent-bit.conf: |
    [SERVICE]
        Flush         1
        Log_Level     info
        Daemon        off
        Parsers_File  parsers.conf
        HTTP_Server   On
        HTTP_Listen   0.0.0.0
        HTTP_Port     2020
    @INCLUDE input-kubernetes.conf
    @INCLUDE filter-kubernetes.conf
    @INCLUDE output-loki.conf
  input-kubernetes.conf: |
    [INPUT]
        Name              tail
        Tag               kube.*
        Path              /var/log/containers/harbor*.log
        Parser            kind
        DB                /var/log/flb_kube.db
        Mem_Buf_Limit     5MB
        Skip_Long_Lines   On
        Refresh_Interval  10
  filter-kubernetes.conf: |
    [FILTER]
        Name                kubernetes
        Match               kube.*
        Kube_URL            https://kubernetes.default.svc:443
        Kube_CA_File        /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        Kube_Token_File     /var/run/secrets/kubernetes.io/serviceaccount/token
        Kube_Tag_Prefix     kube.var.log.containers.
        Merge_Log           On
        Merge_Log_Key       log_processed
        K8S-Logging.Parser  On
        K8S-Logging.Exclude Off
  output-syslog.conf: |
    [OUTPUT]
        Name            syslog
        Match           *
        Host            \${LOG_HOST}
        Port            \${LOG_PORT}
        Mode            udp
        Syslog_Message_Key  message
    [OUTPUT]
        Name  stdout
        Match *
  output-loki.conf: |
    [OUTPUT]
        Name            loki
        Match           *
        Host            \${LOG_HOST}
        Port            \${LOG_PORT}
        auto_kubernetes_labels on
        label_keys      \$kubernetes['pod_name'],\$kubernetes['container_name'],\$kubernetes['namespace_name'],\$kubernetes['host']
    [OUTPUT]
        Name  stdout
        Match *
  parsers.conf: |
    [PARSER]
        Name   apache
        Format regex
        Regex  ^(?<host>[^ ]*) [^ ]* (?<user>[^ ]*) \[(?<time>[^\]]*)\] "(?<method>\S+)(?: +(?<path>[^\"]*?)(?: +\S*)?)?" (?<code>[^ ]*) (?<size>[^ ]*)(?: "(?<referer>[^\"]*)" "(?<agent>[^\"]*)")?$
        Time_Key time
        Time_Format %d/%b/%Y:%H:%M:%S %z
    [PARSER]
        Name   apache2
        Format regex
        Regex  ^(?<host>[^ ]*) [^ ]* (?<user>[^ ]*) \[(?<time>[^\]]*)\] "(?<method>\S+)(?: +(?<path>[^ ]*) +\S*)?" (?<code>[^ ]*) (?<size>[^ ]*)(?: "(?<referer>[^\"]*)" "(?<agent>[^\"]*)")?$
        Time_Key time
        Time_Format %d/%b/%Y:%H:%M:%S %z
    [PARSER]
        Name   apache_error
        Format regex
        Regex  ^\[[^ ]* (?<time>[^\]]*)\] \[(?<level>[^\]]*)\](?: \[pid (?<pid>[^\]]*)\])?( \[client (?<client>[^\]]*)\])? (?<message>.*)$
    [PARSER]
        Name   nginx
        Format regex
        Regex ^(?<remote>[^ ]*) (?<host>[^ ]*) (?<user>[^ ]*) \[(?<time>[^\]]*)\] "(?<method>\S+)(?: +(?<path>[^\"]*?)(?: +\S*)?)?" (?<code>[^ ]*) (?<size>[^ ]*)(?: "(?<referer>[^\"]*)" "(?<agent>[^\"]*)")?$
        Time_Key time
        Time_Format %d/%b/%Y:%H:%M:%S %z
    [PARSER]
        Name   json
        Format json
        Time_Key time
        Time_Format %d/%b/%Y:%H:%M:%S %z
    [PARSER]
        Name        docker
        Format      json
        Time_Key    time
        Time_Format %Y-%m-%dT%H:%M:%S.%L
        Time_Keep   On
    [PARSER]
        Name        kind
        Format      regex
        Regex       ^(?<time>.+) (?<stream>stdout|stderr) (?<logtag>[FP]) (?<message>.*)$
        Time_Key    time
        Time_Format %Y-%m-%dT%H:%M:%S.%L
        Time_Keep   On
    [PARSER]
        # http://rubular.com/r/tjUt3Awgg4
        Name cri
        Format regex
        Regex ^(?<time>[^ ]+) (?<stream>stdout|stderr) (?<logtag>[^ ]*) (?<message>.*)$
        Time_Key    time
        Time_Format %Y-%m-%dT%H:%M:%S.%L%z
    [PARSER]
        Name        syslog
        Format      regex
        Regex       ^\<(?<pri>[0-9]+)\>(?<time>[^ ]* {1,2}[^ ]* [^ ]*) (?<host>[^ ]*) (?<ident>[a-zA-Z0-9_\/\.\-]*)(?:\[(?<pid>[0-9]+)\])?(?:[^\:]*\:)? *(?<message>.*)$
        Time_Key    time
        Time_Format %b %d %H:%M:%S
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluent-bit
  namespace: logging
  labels:
    k8s-app: fluent-bit-logging
    version: v1
    kubernetes.io/cluster-service: "true"
spec:
  selector:
    matchLabels:
      k8s-app: fluent-bit-logging
  template:
    metadata:
      labels:
        k8s-app: fluent-bit-logging
        version: v1
        kubernetes.io/cluster-service: "true"
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "2020"
        prometheus.io/path: /api/v1/metrics/prometheus
    spec:
      containers:
      - name: fluent-bit
        image: ap.cicd.harbor.vmwarecna.net/proxy/fluent/fluent-bit:1.6.10
        imagePullPolicy: IfNotPresent
        ports:
          - containerPort: 2020
        env:
        - name: LOG_HOST
          value: "loki.logging.svc"
        - name: LOG_PORT
          value: "3100"
        volumeMounts:
        - name: varlog
          mountPath: /var/log
        - name: varlibdockercontainers
          mountPath: /var/lib/docker/containers
          readOnly: true
        - name: fluent-bit-config
          mountPath: /fluent-bit/etc/
      terminationGracePeriodSeconds: 10
      volumes:
      - name: varlog
        hostPath:
          path: /var/log
      - name: varlibdockercontainers
        hostPath:
          path: /var/lib/docker/containers
      - name: fluent-bit-config
        configMap:
          name: fluent-bit-config
      serviceAccountName: fluent-bit
      tolerations:
      - key: node-role.kubernetes.io/master
        operator: Exists
        effect: NoSchedule
      - operator: "Exists"
        effect: "NoExecute"
      - operator: "Exists"
        effect: "NoSchedule"
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: grafana
  namespace: logging
spec:
  rules:
  - host: grafana.harbor.domain
    http:
      paths:
      - path: /
        backend:
          serviceName: grafana
          servicePort: 3000
EOF
```

- you can use port-forward to access `http://grafana.harbor.domain` after set /etc/hosts

```bash
sudo kubectl port-forward svc/ingress-nginx-controller -n ingress-nginx --address 0.0.0.0 80:80 443:443
```

- login and add datasource loki `http://loki.logging.svc:3100`

- you can filter pod by label `kubernetes_container_name`

- you can collect all other pods' logs by changing `/var/log/containers/*.log`
