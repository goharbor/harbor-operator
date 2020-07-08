package trivy

import (
	"context"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
)

func (r *Reconciler) GetConfigMap(ctx context.Context, trivy *goharborv1alpha2.Trivy) (*corev1.ConfigMap, error) {
	name := r.NormalizeName(ctx, trivy.GetName())
	namespace := trivy.GetNamespace()

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		// populate
		Data: map[string]string{
			"SCANNER_LOG_LEVEL":                     trivy.Spec.Log.LogLevel,
			"SCANNER_API_SERVER_ADDR":               trivy.Spec.Server.Address,
			"SCANNER_API_SERVER_TLS_CERTIFICATE":    trivy.Spec.Server.HTTPS.CertificateRef,
			"SCANNER_API_SERVER_TLS_KEY":            trivy.Spec.Server.HTTPS.KeyRef,
			"SCANNER_API_SERVER_CLIENT_CAS":         trivy.Spec.Server.HTTPS.ClientCas,
			"SCANNER_API_SERVER_READ_TIMEOUT":       trivy.Spec.Server.ReadTimeout.Duration.String(),
			"SCANNER_API_SERVER_WRITE_TIMEOUT":      trivy.Spec.Server.WriteTimeout.Duration.String(),
			"SCANNER_API_SERVER_IDLE_TIMEOUT":       trivy.Spec.Server.IdleTimeout.Duration.String(),
			"SCANNER_TRIVY_CACHE_DIR":               trivy.Spec.Server.CacheDir,
			"SCANNER_TRIVY_REPORTS_DIR":             trivy.Spec.Server.ReportsDir,
			"SCANNER_TRIVY_DEBUG_MODE":              strconv.FormatBool(trivy.Spec.Server.DebugMode),
			"SCANNER_TRIVY_VULN_TYPE":               strings.Join(trivy.Spec.Server.VulnType, ","),
			"SCANNER_TRIVY_SEVERITY":                strings.Join(trivy.Spec.Server.Severity, ","),
			"SCANNER_TRIVY_IGNORE_UNFIXED":          strconv.FormatBool(trivy.Spec.Server.IgnoreUnfixed),
			"SCANNER_TRIVY_SKIP_UPDATE":             strconv.FormatBool(trivy.Spec.Server.SkipUpdate),
			"SCANNER_TRIVY_GITHUB_TOKEN":            trivy.Spec.Server.GithubToken,
			"SCANNER_TRIVY_INSECURE":                strconv.FormatBool(trivy.Spec.Server.Insecure),
			"SCANNER_STORE_REDIS_NAMESPACE":         trivy.Spec.Cache.Redis.DSN,
			"SCANNER_STORE_REDIS_SCAN_JOB_TTL":      trivy.Spec.Cache.RedisScanJobTTL.Duration.String(),
			"SCANNER_JOB_QUEUE_REDIS_NAMESPACE":     trivy.Spec.Cache.QueueRedisNamespace,
			"SCANNER_JOB_QUEUE_WORKER_CONCURRENCY":  strconv.Itoa(trivy.Spec.Cache.QueueWorkerConcurrency),
			"SCANNER_REDIS_POOL_MAX_ACTIVE":         strconv.Itoa(trivy.Spec.Cache.PoolMaxActive),
			"SCANNER_REDIS_POOL_MAX_IDLE":           strconv.Itoa(trivy.Spec.Cache.PoolMaxIdle),
			"SCANNER_REDIS_POOL_IDLE_TIMEOUT":       trivy.Spec.Cache.PoolIdleTimeout.Duration.String(),
			"SCANNER_REDIS_POOL_CONNECTION_TIMEOUT": trivy.Spec.Cache.PoolConnectionTimeout.Duration.String(),
			"SCANNER_REDIS_POOL_READ_TIMEOUT":       trivy.Spec.Cache.PoolReadTimeout.Duration.String(),
			"SCANNER_REDIS_POOL_WRITE_TIMEOUT":      trivy.Spec.Cache.PoolWriteTimeout.Duration.String(),
		},
	}, nil
}
