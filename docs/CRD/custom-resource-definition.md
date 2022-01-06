# Custom Resource Definition

This document guides user to learn the related fields defined in the `HarborCluster` CRD and then customize their Harbor cluster deployment stack.

**CRD version**: `v1beta1`

## CRD spec

Describe the spec fields with YAML code snippets and comments.
All the parts here share the head YAML code snippet shown below.

```yaml
apiVersion: goharbor.io/v1beta1
kind: HarborCluster
metadata:
  name: harborcluster-sample
  namespace: cluster-sample-ns
spec:
  # ... Skipped fields
  # ...
  # ... Skipped fields
```

### Top level general fields

`expose`(required): Expose the access endpoints of Harbor core services as well as notary service (optional).

```yaml
spec:
  # ... Skipped fields
  
  expose:
    # Expose core services
    core: # Required
      # TLS setting
      tls: # Optional
        # Certificate reference
        certificateRef: <cert-ref> # Optional
      # Expose service with ingress way
      ingress:
        # Host of the exposed service
        host: <registry.goharbor.io> # Required
        # Ingress controller type, support ["gce","ncp","contour","default"]
        # "default" means nginx
        controller: default # Optional, default value = "default"
        # Annotations applied to the ingress
        annotations: # Optional
          key: value
        # Set the ingress class name. If it is not set, the system default one will be picked up.
        ingressClassName: ingressClass # Optional
    # Expose notary service when it is configured
    notary: # Optional
      ## Totally same with above [expose.core] part, skipped here.

  # ... Skipped fields
```

`externalURL`(required): the public URL with pattern `https?://.*` for accessing Harbor registry.

```yaml
spec:

  # ... Skipped fields

  externalURL: https://registry.goharbor.io # Required

  # ... Skipped fields
```

`internalTLS`(optional): enable secure communications between Harbor components if it is set.

```yaml
spec:
  # ... Skipped fields

  internalTLS: # Optional
    enabled: true # Optional, default = false

  # ... Skipped fields
```

`logLevel`(optional): set the log level of the Harbor loggers.

```yaml
spec:
  # ... Skipped fields
  
  # Support settings ["debug","info","warning","error","fatal"]
  logLevel: "debug" # Optional, default = "info"

  # ... Skipped fields
```

`harborAdminPasswordRef`(required): the secret reference containing the preset admin password.

```yaml
spec:
  # ... Skipped fields
  
  harborAdminPasswordRef: "myAdminPwd" # Required
  
  # ... Skipped fields
```

`updateStrategyType`(optional): the update strategy.

```yaml
spec:
  # ... Skipped fields

  updateStrategyType: "RollingUpdate" # Optional, default="RollingUpdate"
  
  # ... Skipped fields
```

`version`(optional): keep the version of the Harbor deployed to cluster. It's mainly used in the version upgrading case.

```yaml
spec:
  # ... Skipped fields
  
  # Example: 2.4.0
  version: <Harbor version> # Optional
  
  # ... Skipped fields
```

`proxy`(optional): configure proxy settings for related Harbor components.

```yaml
spec:
  # ... Skipped fields
  
  proxy: # Optional
    # HTTP proxy
    httpProxy: http://my.proxy.com:8888 # Optional, pattern="https?://.+"
    # HTTPS proxy
    httpsProxy: http://my.proxy.com:8888 # Optional, pattern="https?://.+"
    # No proxy
    noProxy: # Optional, default=["127.0.0.1","localhost",".local",".internal"]
      - 127.0.0.1
      - localhost
      - .local
      - .internal
    # Configure proxy settings for which components
    components: # Optional, default=[core,jobservice,trivy]
      - core
      - jobservice
      - trivy
  
  # ... Skipped fields
```

### Configure image source

`imageSource` configures the general image source from where pulling images. Image settings configured here are applicable to all the components.

```yaml
spec:
  # ... Skipped fields
  
  imageSource: # Optional
    # The root repository path of the component images.
    # e.g: if it is set to 'docker.io/goharbor', then the core image path will be 'docker.io/goharbor/harbor-core'
    repository: docker.io/goharbor # Required
    # The tag suffix of the component images.
    # e.g: if it is set to `-staging`, then the core image path will be 'docker.io/goharbor/harbor-core:<version>-staging'
    tagSuffix: -staging # Optional
    # Image pull policy. Support values are ["Always","Never","IfNotPresent"].
    # More info: https://kubernetes.io/docs/concepts/containers/images#updating-images
    imagePullPolicy: Always # Optional, default = IfNotPresent
    # Image pull secrets
    imagePullSecrets: # Optional
      - name: myHarborRegSecret
  
  # ... Skipped fields
```

### Network stack settings

Network settings for the deploying Harbor.

```yaml
spec:
  # ... Skipped fields
  
  # Network settings
  network: # Optional
    # Set what IP families are used for the deploying Harbor
    ipFamilies:
      - IPv4
      - IPv6
  
  # ... Skipped fields
```

### Trace settings

Tracing settings for the deploying Harbor.

```yaml
spec:
  # ... Skipped fields
  
  # Tracing settings
  trace: # Optional
    # Enable tracing or not
    enabled: false # Optional, default is false
    # Used to differentiate different harbor services.
    namespace: core # Optional
    # Set `sampleRate` to 1 if you wanna sampling 100% of trace data; set 0.5 if you wanna sampling 50% of trace data, and so forth.
    sampleRate: 1 # Optional, default is 1
    # A key value dict contains user defined attributes used to initialize trace provider.
    attributes: # Optional
      key: value
    # The tracing provider: 'jaeger' or 'otel'
    provider: jaeger # Required
    # Spec for jaeger provider if provider is set to jaeger
    jaeger: # Optional
      # Serve mode. `collector` or `agent`
      mode: collector # Required
      # Configuration for collector mode
      collector: # Optional
        # The endpoint of the jaeger collector
        endpoint: jaeger.io # Required
        # The username of the jaeger collector
        username: foo
        # The password secret reference name of the jaeger collector
        passwordRef: foobar
      # Configuration for agent mode
      agent: # Optional
        # The host of the jaeger agent
        host: jaeger.io # Required
        # The port of the jaeger agent
        port: 8000
    # Spec for otel provider if provider is set to otel.
    otel: # Optional
      # The endpoint of otel
      endpoint: otel.io # Required
      # The URL path of otel
      urlPath: /otel # Required
      # Whether enable compression or not for otel
      compression: false # Optional
      # Whether establish insecure connection or not for otel
      insecure: true # Optional
      # The timeout of otel
      timeout: 10s # Optional, default is 10s

  # ... Skipped fields
```

### Harbor component related fields

Each Harbor component has its own spec to accept configurations and shares the common spec shown below.

```yaml
spec:
 # ... Skipped fields

  # Besides the common component spec, no extra parts for component 'portal'
  portal:
    # Common component spec

    # Image name for the component. It will override the default one.
    image: my-portal # Optional
    # Image pull policy. It will override the global 'imageSource' settings and the default one.
    imagePullPolicy: # Optional, default = IfNotPresent
    # Image pull secrets. It will override the global 'imageSource' settings if it has been set.
    imagePullSecrets:
      - name: myHarborRegSecretOfPortal

    # Replicas is the number of desired replicas.
    # This is a pointer to distinguish between explicit zero and unspecified.
    # More info: https://kubernetes.io/docs/concepts/workloads/controllers/replicationcontroller#what-is-a-replicationcontroller
    replicas: 3 # Optional

    # ServiceAccountName is the name of the ServiceAccount to use to run this component.
    # More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/
    serviceAccountName: mySA # Optional

    # NodeSelector is a selector which must be true for the component to fit on a node.
    # Selector which must match a node's labels for the pod to be scheduled on that node.
    # More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
    nodeSelector: # Optional
      key: value

    # If specified, the pod's tolerations.
    # More info: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
    tolerations: {} # Optional

    # Compute Resources required by this component.
    # Cannot be updated.
    # More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
    resources: # Optional
      # Limits describes the maximum amount of compute resources allowed.
      # More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
      limits: {} # Optional
      # Requests describes the minimum amount of compute resources required.
      # If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
      # otherwise to an implementation-defined value.
      # More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
      requests: {} # Optional

  # The following components also includes the common spec shown above.
  # ... Skip duplicated configurations here
  core: {}
  jobservice: {}
  registry: {}
  registryctl: {}
  chartmuseum: {}
  trivy: {}
  exporter: {}
  notary:
    server: {}
    signer: {}
  
 # ... Skipped fields
```

Extra configurations for Harbor component `core`.

```yaml
spec:
  # ... Skipped fields
  
  core:
    # ... Skipped common component spec here

    # Extra configurations

    # Certificates need to be injected into core
    certificateRefs: # Optional
      - cert1
      - cert2
    # Token issuer
    tokenIssuer: myIssuer # Required
    # Metrics settings
    metrics: # optional
      enabled: false # optional, default is false
      port: 8001 # optional, default is 8001
      path: /metrics # optional, default is /metrics
  
  # ... Skipped fields
```

Extra configurations for Harbor component `jobservice`.

```yaml
spec:
  # ... Skipped fields

  jobservice:
    # ... Skipped common component spec here

    # Extra configurations

    # Certificates need to be injected into jobservice
    certificateRefs: # Optional
      - cert1
      - cert2
    # The number of workers
    workerCount: 10 # Optional, default = 10 , minimal = 1
    # Metrics settings
    metrics: # optional
      # Similar to the section shown in the `core` component
      # Skip here
  
  # ... Skipped fields
```

Extra configurations for Harbor component `registry`.

```yaml
spec:
  # ... Skipped fields

  registry:
    # ... Skipped common component spec here

    # Extra configurations

    # Enable relative URLs
    relativeURLs: true # Optional, default = true
    # Middlewares for storage
    storageMiddlewares: # Optional
      - name: m1 # Required
        optionsRef: op1 # Optional
    metrics: # optional
    # Similar to the section shown in the `core` component
    # Skip here
  
  # ... Skipped fields
```

Extra configurations for Harbor component `chartmuseum`.

```yaml
spec:
  # ... Skipped fields

  chartmuseum:
    # ... Skipped common component spec here

    # Extra configurations

    # Certificates need to be injected into chartmuseum
    certificateRefs: # Optional
      - cert1
      - cert2
    # Harbor defaults ChartMuseum to returning relative URLs,
    # If you want using absolute URL you should enable it.
    absoluteUrl: false # Optional, default = false
  
  # ... Skipped fields
```

Extra configurations for Harbor component `trivy`.

```yaml
spec:
  # ... Skipped fields

  trivy:
    # ... Skipped common component spec here

    # Extra configurations

    # Certificates need to be injected into chartmuseum
    certificateRefs: # Optional
      - cert1
      - cert2
    # The name of the secret containing the token to connect to GitHub API.
    githubTokenRef: github-token # Optional
    # The flag to enable or disable Trivy DB downloads from GitHub
    skipUpdate: false # Optional, default = false
    # Storage used for keep data by trivy.
    storage: # required
      # ReportsPersistentVolume specify the persistent volume used to store Trivy reports.
      reportsPersistentVolume: # Optional, if it is not set, then empty dir will be used.
        # Inline the corev1.PersistentVolumeClaimVolumeSource
        # ClaimName is the name of a PersistentVolumeClaim in the same namespace as the pod using this volume.
        # More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims
        claimName: myPVC # Required
        # Will force the ReadOnly setting in VolumeMounts.
        readOnly: false # Optional
        prefix: myPrefix # Optional
      # CachePersistentVolume specify the persistent volume used to store Trivy cache.
      # Same configurations with ReportsPersistentVolume.
      cachePersistentVolume: {} # Optional, if it is not set, then empty dir will be used.
  
  # ... Skipped fields
```

Extra configurations for Harbor component `notary`.

```yaml
spec:
  # ... Skipped fields

  notary:
    server: {} # Skipped common component spec here ...
    signer: {} # Skipped common component spec here ...

    # Extra configurations

    # Inject migration configuration to notary resources
    migrationEnabled: true # Optional, default = true
  
  # ... Skipped fields
```

### Storage related fields

So far, there are 6 options for storage configurations: `FileSystem` ([Persistent Volume](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)), [S3](https://docs.aws.amazon.com/AmazonS3/latest/API/Welcome.html) , [Swift](https://docs.openstack.org/swift/latest/), [Azure](https://azure.microsoft.com/services/storage/), [Gcs](https://cloud.google.com/storage) and MinIO.

#### FileSystem

Configure `filesystem` as backend storage.

```yaml
spec:
  # ... Skipped fields
  
  storage:
    kind: "FileSystem"
    spec:
      # FileSystem is an implementation of the storagedriver.StorageDriver interface which uses the local filesystem.
      # The local filesystem can be a remote volume.
      # See: https://docs.docker.com/registry/storage-drivers/filesystem/
      filesystem: # Optional
        chartPersistentVolume: # Optional
          # Inline the corev1.PersistentVolumeClaimVolumeSource
          # ClaimName is the name of a PersistentVolumeClaim in the same namespace as the pod using this volume.
          # More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims
          claimName: myPVC # Required
          # Will force the ReadOnly setting in VolumeMounts.
          readOnly: false # Optional
          prefix: myPrefix # Optional
        registryPersistentVolume: # Optional
          # ... Skipped the same fields with 'chartPersistentVolume': 'claimName', 'readOnly' and 'prefix'.
          # ...
          # Max threads
          maxthreads: 100 # Optional, default = 100, minimal = 25

  # ... Skipped fields
```

#### S3

Configure `s3` as backend storage.

```yaml
spec:
  # ... Skipped fields

  storage:
    kind: "S3"
    spec:
      # Configure S3 as the backend storage of Harbor.
      # An implementation of the storagedriver.StorageDriver interface which uses Amazon S3 or S3 compatible services for object storage.
      # See: https://docs.docker.com/registry/storage-drivers/s3/
      s3: # Optional
        # The AWS Access Key.
        # If you use IAM roles, omit to fetch temporary credentials from IAM.
        accesskey: ak # Optional
        # Reference to the secret containing the AWS Secret Key.
        # If you use IAM roles, omit to fetch temporary credentials from IAM.
        secretkeyRef: secret # Optional
        # The AWS region in which your bucket exists.
        # For the moment, the Go AWS library in use does not use the newer DNS based bucket routing.
        # For a list of regions, see http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html
        region: us-east-1 # Required
        # Endpoint for S3 compatible storage services (Minio, etc).
        regionendpoint: Minio # Required
        # The bucket name in which you want to store the registry’s data.
        bucket: default # Required
        # This is a prefix that is applied to all S3 keys to allow you to segment data in your bucket if necessary.
        rootdirectory: registry # Optional
        # The S3 storage class applied to each registry file.
        storageclass: STANDARD # Optional, default="STANDARD"
        # KMS key ID to use for encryption (encrypt must be true, or this parameter is ignored).
        keyid: kid # Optional
        # Specifies whether the registry stores the image in encrypted format or not. A boolean value.
        encrypt: false # Optional, default=false
        # Skips TLS verification when the value is set to true.
        skipverify: false # Optional, default=false
        # Certificate
        certificateRef: cert # Optional
        # Indicates whether to use HTTPS instead of HTTP. A boolean value.
        secure: true # Optional, default=true
        # Indicates whether the registry uses Version 4 of AWS’s authentication.
        v4auth: true # Optional, default=true
        # The S3 API requires multipart upload chunks to be at least 5MB.
        chunksize: 5242880 # Optional, minimal = 5242880

  # ... Skipped fields
```

#### Swift

Configure `swift` as backend storage.
This method is not recommended since Swift is *enventual consistent*.
Please use [S3 Middleware](https://docs.openstack.org/newton/config-reference/object-storage/configure-s3.html) in front of Swift and configure 2nd method: S3 storage.

```yaml
spec:
  # ... Skipped fields

  storage:
    kind: Swift
    spec:
      # Configure Swift as the backend storage of Harbor.
      # An implementation of the storagedriver.StorageDriver interface that uses OpenStack Swift for object storage.
      # See: https://docs.docker.com/registry/storage-drivers/swift/

      swift: # Optional
        # URL for obtaining an auth token.
        # https://storage.myprovider.com/v2.0 or https://storage.myprovider.com/v3/auth
        authurl: https://storage.myprovider.com/v3/auth # Required
        # The Openstack user name.
        username: openstack-user # Required
        # Secret name containing the Openstack password.
        passwordRef: secret # Required
        # The Openstack region in which your container exists.
        region: region-1 # Optional
        # The name of your Swift container where you wish to store the registry’s data.
        # The driver creates the named container during its initialization.
        container: container1 # Required
        # You can either use tenant or tenantid.
        tenant: myTenant # Optional
        # You can either use tenant or tenantid.
        tenantid: myTenantID # Optional
        # Your Openstack domain name for Identity v3 API. You can either use domain or domainid.
        domain: sampleDomain # Optional
        # Your Openstack domain ID for Identity v3 API. You can either use domain or domainid.
        domainid: did # Optional
        # Your Openstack trust ID for Identity v3 API.
        trustid: myTrustID # Optional
        # Skips TLS verification if the value is set to true.
        insecureskipverify: false # Optional, default=false
        # Size of the data segments for the Swift Dynamic Large Objects.
        # This value should be a number.
        chunksize: 5242880 # Optional, minimal=5242880
        # This is a prefix that is applied to all Swift keys to allow you to segment data in your container if necessary. Defaults to the container’s root.
        prefix: registry # Optional
        # The secret key used to generate temporary URLs.
        secretkeyRef: key # Optional
        # The access key to generate temporary URLs. It is used by HP Cloud Object Storage in addition to the secretkey parameter.
        accesskey: ak # Optional
        # Specify the OpenStack Auth’s version, for example 3. By default the driver autodetects the auth’s version from the authurl.
        authversion: 3 # Optional
        # The endpoint type used when connecting to swift.
        # Supports values ["public","internal","admin"]
        endpointtype: public # Optional, default=public

  # ... Skipped fields
```

#### MinIO

The minio storage configuration can be configured to let the Harbor operator automatically deploy an in-cluster S3 compatible  service with HA supported as the backend storage service of the deploying Harbor.

```yaml
spec:
  # ... Skipped fields
  
  storage:
    # Kind of which storage service to be used. Only support MinIO now.
    kind: MinIO # Required
    # MinIO configurations
    spec: # Required
      # Image name for the MinIO. It will override the default one.
      image: my-minio # Optional
      # Image pull policy. It will override the global 'imageSource' settings and the default one.
      imagePullPolicy: # Optional, default = IfNotPresent
      # Image pull secrets. It will override the global 'imageSource' settings if it has been set.
      imagePullSecrets:
        - name: myHarborRegSecretOfMinIO
      # Redirection configurations.
      redirect: # Required
        # Determine if the redirection of minio storage is enabled.
        enable: true # Optional, default=true
        # If 'enable' is set to be true, then configure extra settings here.
        # Expose MinIO service for client accessing.
        # Same configuration with the top level 'expose' section.
        expose: # Optional
          # TLS setting
          tls: # Optional
            # Certificate reference
            certificateRef: <cert-ref> # Optional
          # Expose service with ingress way
          ingress:
            # Host of the exposed service
            host: <minio.goharbor.io> # Required
            # Ingress controller type, support ["gce","ncp","default"]
            controller: default # Optional, default value = "default"
            # Annotations applied to the ingress
            annotations: # Optional
              key: value
      # Reference to the secret containing the MinIO access key and secret key.
      secretRef: minioSecret # Optional
      # Supply number of replicas.
      # For standalone mode, supply 1. For distributed mode, supply 4 or more.
      # Note that the operator does not support upgrading from standalone to distributed mode.
      # Specially, 'replicas'*'volumesPerServer' should be >=4.
      replicas: 4 # Required, minimal=1
      # Number of persistent volumes that will be attached per server
      # Specially, 'replicas'*'volumesPerServer' should be >=4.
      volumesPerServer: 1 # Required, minimal=1
      # VolumeClaimTemplate allows a user to specify how volumes inside a MinIOInstance.
      # More info: https://github.com/kubernetes/kubernetes/blob/6b1d87acf3c8253c123756b9e61dac642678305f/pkg/apis/core/types.go#L398
      volumeClaimTemplate: {} # Optional
      # If provided, use these requests and limit for cpu/memory resource allocation
      # More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
      resources: # Optional
        # Limits describes the maximum amount of compute resources allowed.
        # More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
        limits: {} # Optional
        # Requests describes the minimum amount of compute resources required.
        # If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
        # otherwise to an implementation-defined value.
        # More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
        requests: {} # Optional

  # ... Skipped fields
```

### Database related fields

Two alternatives provided to configure the database service used by the deploying Harbor.

#### Standard Database

Standard database configurations can be used to set the *existing pre-deployed* or *cloud database services* as the dependent database of the deploying Harbor.

```yaml
spec:
  # ... Skipped fields

  # Configure existing pre-deployed or cloud database service.
  database:
    kind: PostgreSQL
    spec:
      postgresql:
        # PostgreSQL user name to connect as.
        # Defaults to be the same as the operating system name of the user running the application.
        username: postsql # Required
        # Secret containing the password to be used if the server demands password authentication.
        passwordRef: psqlSecret # Optional
        # PostgreSQL hosts.
        # At least 1.
        hosts:
          # Name of host to connect to.
          # If a host name begins with a slash, it specifies Unix-domain communication rather than
          # TCP/IP communication; the value is the name of the directory in which the socket file is stored.
          - host: psql # Required
          # Port number to connect to at the server host,
          # or socket file name extension for Unix-domain connections.
          # Zero, specifies the default port number established when PostgreSQL was built.
            port: 5432 # Optional
        # PostgreSQL has native support for using SSL connections to encrypt client/server communications for increased security.
        # Supports values ["disable","allow","prefer","require","verify-ca","verify-full"].
        sslMode: prefer # Optional, default=prefer
        prefix: prefix # Optional
  
  # ... Skipped fields
```

#### in-cluster database configuration

The *in-cluster database configuration* can be configured to let the Harbor operator automatically deploy an in-cluster PostgreSQL database service with HA supported as the dependent database of the deploying Harbor.

```yaml
spec:
  # ... Skipped fields

  # database configurations.
  database:
    # Set the kind of which database service to be used.
    kind: "Zlando/PostgreSQL" # Required
    # storage spec
    spec: # Required
      zlandoPostgreSql:
        # Image name for the PostgresSQL. It will override the default one.
        image: my-psql # Optional
        # Image pull policy. It will override the global 'imageSource' settings and the default one.
        imagePullPolicy: # Optional, default = IfNotPresent
        # Image pull secrets. It will override the global 'imageSource' settings if it has been set.
        imagePullSecrets:
          - name: myHarborRegSecretOfPsql
        # Specify the storage size for the PostgresSQL.
        storage: 1Gi # Optional, default="1Gi"
        # Replicas of PostgresSQL instances.
        replicas: 3 # Optional, default=3
        # The storage class used for creating storage.
        storageClassName: default # Optional
        # If provided, use these requests and limit for cpu/memory resource allocation
        # More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
        resources: # Optional
          # Limits describes the maximum amount of compute resources allowed.
          # More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
          limits: {} # Optional
          # Requests describes the minimum amount of compute resources required.
          # If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
          # otherwise to an implementation-defined value.
          # More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
          requests: {} # Optional


  # ... Skipped fields
```

### Cache related fields

Two alternatives provided to configure the cache(`Redis`) service used by the deploying Harbor.

#### Standard Cache

Standard cache configurations can be used to set the *existing pre-deployed* or *cloud cache services* as the dependent cache of the deploying Harbor.

```yaml
spec:
  # ... Skipped fields

  # Cache configuration.
  cache: # Optional
    kind: "Redis"
    spec:
      redis:
        # Server host.
        host: myredis.com # Required
        # Server port.
        port: 6347 # Required
        # For setting sentinel masterSet.
        sentinelMasterSet: sentinel # Optional
        # Secret containing the password to use when connecting to the server.
        passwordRef: pwdSecret # Optional
        # Secret containing the client certificate to authenticate with.
        certificateRef: cert # Optional
  
  # ... Skipped fields
```

#### in-cluster cache configuration

The *in-cluster cache configuration* can be configured to let the Harbor operator automatically deploy an in-cluster Redis service with HA supported as the dependent cache of the deploying Harbor.

```yaml
spec:
  # ... Skipped fields

  # cache configurations.
  cache:
    # Set the kind of cache service to be used. Only support 'Redis' now.
    kind: RedisFailover # Required
    # Redis configuration spec.
    spec: # Required
      redisFailover:
        # Image name for the Redis. It will override the default one.
        image: my-redis # Optional
        # Image pull policy. It will override the global 'imageSource' settings and the default one.
        imagePullPolicy: # Optional, default = IfNotPresent
        # Image pull secrets. It will override the global 'imageSource' settings if it has been set.
        imagePullSecrets:
          - name: myHarborRegSecretOfRedis
        # Redis sentinel
        sentinel: # Required
          # Replicas of the sentinel service.
          replicas: 3 # Optional, default=3
        # Redis server.
        server: # Required
          # Replicas of the server.
          replicas: 3 # Optional, default=3
          # Storage class used to apply storage of redis.
          StorageClassName: default # Optional
          # Storage size.
          storage: 1Gi # Optional
          # If provided, use these requests and limit for cpu/memory resource allocation
          # More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
          resources: # Optional
            # Limits describes the maximum amount of compute resources allowed.
            # More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
            limits: {} # Optional
            # Requests describes the minimum amount of compute resources required.
            # If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
            # otherwise to an implementation-defined value.
            # More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
            requests: {} # Optional

  # ... Skipped fields

```

## Status spec

The status spec of the CR `HarborCluster` is described as below:

```yaml
status:
  # Show the versioning info of the running operator.
  operator:
    # The code commit for building the running operator.
    controllerGitCommit: <commit_hash> # Optional
    # The version of the running operator.
    controllerVersion: 1.0.0 # Optional
    # Name of the operator controller.
    controllerName: harbor
  # Overall status of HarborCluster CR.
  # Status can be "creating", "healthy" and "unhealthy"
  status: healthy
  # Condition list
  conditions:
    # The type of the condition.
    # It can be "ServiceReady", "StorageReady", "DatabaseReady", "CacheReady" and "ConfigurationReady".
    type: ServiceReady
    # Status is the status of the condition.
    # Can be True, False, Unknown.
    status: True
    # Last time the condition transitioned from one status to another.
    lastTransitionTime: <time> # Optional
    # Unique, one-word, CamelCase reason for the condition's last transition.
    reason: "reason" # Optional
    # Human-readable message indicating details about last transition.
    message: "message" # Optional
```
