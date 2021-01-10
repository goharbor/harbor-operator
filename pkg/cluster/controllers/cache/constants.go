package cache

const (
	ErrorGetRedisClient               = "Get redis client error"
	ErrorCheckRedisHealth             = "Check redis health error"
	ErrorCreateComponentSecret        = "Create component secret error"
	ErrorGenerateRedisCr              = "Generate redis cr error"
	ErrorSetOwnerReference            = "Set owner reference error"
	ErrorCreateRedisSecret            = "Create redis secret error"
	ErrorCreateRedisCr                = "Create redis cr error"
	ErrorCreateRedisServerService     = "Create redis server service error"
	ErrorGetRedisCr                   = "Get redis cr error"
	ErrorGetRedisServerPod            = "Get redis server pod error"
	ErrorGetRedisSentinelPod          = "Get redis sentinel pod error"
	ErrorGetRedisPassword             = "Get redis password error"
	ErrorCheckRedisIsMaster           = "Check redis isMaster error"
	ErrorManualFailoverRedis          = "Manual failover redis error"
	ErrorUpdateRedisCr                = "Update redis cr error"
	ErrorDefaultUnstructuredConverter = "Default unstructured converter error"
)
