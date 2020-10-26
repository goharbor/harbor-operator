package cache

const (
	GetRedisClientError               = "Get redis client error"
	CheckRedisHealthError             = "Check redis health error"
	CreateComponentSecretError        = "Create component secret error"
	GenerateRedisCrError              = "Generate redis cr error"
	SetOwnerReferenceError            = "Set owner reference error"
	CreateRedisSecretError            = "Create redis secret error"
	CreateRedisCrError                = "Create redis cr error"
	CreateRedisServerServiceError     = "Create redis server service error"
	GetRedisCrError                   = "Get redis cr error"
	GetRedisServerPodError            = "Get redis server pod error"
	GetRedisSentinelPodError          = "Get redis sentinel pod error"
	GetRedisPasswordError             = "Get redis password error"
	CheckRedisIsMasterError           = "Check redis isMaster error"
	ManualFailoverRedisError          = "Manual failover redis error"
	UpdateRedisCrError                = "Update redis cr error"
	DefaultUnstructuredConverterError = "Default unstructured converter error"
)

const (
	RedisSentinelSchema = "sentinel"
	RedisServerSchema   = "redis"
)

const (
	ExternalComponent  string = "external"
	InClusterComponent string = "inCluster"
)
