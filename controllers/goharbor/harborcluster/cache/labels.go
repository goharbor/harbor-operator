package cache

const (
	AppLabel = "goharbor.io/harbor-cluster"
)

// NewLabels returns new labels
func (redis *RedisReconciler) NewLabels() map[string]string {
	dynLabels := map[string]string{
		"app.kubernetes.io/name":     "cache",
		"app.kubernetes.io/instance": redis.HarborCluster.Namespace,
		AppLabel:                     redis.HarborCluster.Name,
	}

	return MergeLabels(redis.Labels, dynLabels, redis.HarborCluster.Labels)
}

// MergeLabels merge new label to existing labels
func MergeLabels(allLabels ...map[string]string) map[string]string {
	res := map[string]string{}

	for _, labels := range allLabels {
		for k, v := range labels {
			res[k] = v
		}
	}
	return res
}
