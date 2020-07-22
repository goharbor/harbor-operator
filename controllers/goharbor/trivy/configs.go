package trivy

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
)

func (r *Reconciler) AddConfigMap(ctx context.Context, trivy *goharborv1alpha2.Trivy) error {
	// Forge the ConfigMap resource
	cm, err := r.GetConfigMap(ctx, trivy)
	if err != nil {
		return errors.Wrap(err, "cannot get config map")
	}

	// Add config map to reconciler controller
	_, err = r.Controller.AddConfigMapToManage(ctx, cm)
	if err != nil {
		return errors.Wrapf(err, "cannot manage config map %s", cm.GetName())
	}

	return nil
}

// Get the config map linked to the trivy deployment.
func (r *Reconciler) GetConfigMap(ctx context.Context, trivy *goharborv1alpha2.Trivy) (*corev1.ConfigMap, error) {
	name := r.NormalizeName(ctx, trivy.GetName())
	namespace := trivy.GetNamespace()

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},

		Data: map[string]string{
			"SCANNER_LOG_LEVEL": string(trivy.Spec.Log.Level),

			"SCANNER_API_SERVER_ADDR":            trivy.Spec.Server.Address,
			"SCANNER_API_SERVER_TLS_CERTIFICATE": trivy.Spec.Server.HTTPS.CertificateRef,
			"SCANNER_API_SERVER_TLS_KEY":         trivy.Spec.Server.HTTPS.KeyRef,
			"SCANNER_API_SERVER_CLIENT_CAS":      trivy.Spec.Server.HTTPS.ClientCas,
			"SCANNER_API_SERVER_READ_TIMEOUT":    trivy.Spec.Server.ReadTimeout.Duration.String(),
			"SCANNER_API_SERVER_WRITE_TIMEOUT":   trivy.Spec.Server.WriteTimeout.Duration.String(),
			"SCANNER_API_SERVER_IDLE_TIMEOUT":    trivy.Spec.Server.IdleTimeout.Duration.String(),

			"SCANNER_TRIVY_CACHE_DIR":      trivy.Spec.Server.CacheDir,
			"SCANNER_TRIVY_REPORTS_DIR":    trivy.Spec.Server.ReportsDir,
			"SCANNER_TRIVY_DEBUG_MODE":     strconv.FormatBool(trivy.Spec.Server.DebugMode),
			"SCANNER_TRIVY_VULN_TYPE":      GetVulnerabilities(trivy.Spec.Server.VulnType),
			"SCANNER_TRIVY_SEVERITY":       GetSeverities(trivy.Spec.Server.Severity),
			"SCANNER_TRIVY_IGNORE_UNFIXED": strconv.FormatBool(trivy.Spec.Server.IgnoreUnfixed),
			"SCANNER_TRIVY_SKIP_UPDATE":    strconv.FormatBool(trivy.Spec.Server.SkipUpdate),
			"SCANNER_TRIVY_GITHUB_TOKEN":   trivy.Spec.Server.GithubToken,
			"SCANNER_TRIVY_INSECURE":       strconv.FormatBool(trivy.Spec.Server.Insecure),

			"SCANNER_STORE_REDIS_NAMESPACE":    trivy.Spec.Cache.RedisNamespace,
			"SCANNER_STORE_REDIS_SCAN_JOB_TTL": trivy.Spec.Cache.RedisScanJobTTL.Duration.String(),

			"SCANNER_JOB_QUEUE_REDIS_NAMESPACE":    trivy.Spec.Cache.QueueRedisNamespace,
			"SCANNER_JOB_QUEUE_WORKER_CONCURRENCY": strconv.Itoa(trivy.Spec.Cache.QueueWorkerConcurrency),

			"SCANNER_REDIS_POOL_IDLE_TIMEOUT":       trivy.Spec.Cache.PoolIdleTimeout.Duration.String(),
			"SCANNER_REDIS_POOL_CONNECTION_TIMEOUT": trivy.Spec.Cache.PoolConnectionTimeout.Duration.String(),
			"SCANNER_REDIS_POOL_READ_TIMEOUT":       trivy.Spec.Cache.PoolReadTimeout.Duration.String(),
			"SCANNER_REDIS_POOL_WRITE_TIMEOUT":      trivy.Spec.Cache.PoolWriteTimeout.Duration.String(),
			"SCANNER_REDIS_POOL_MAX_ACTIVE":         strconv.Itoa(trivy.Spec.Cache.PoolMaxActive),
			"SCANNER_REDIS_POOL_MAX_IDLE":           strconv.Itoa(trivy.Spec.Cache.PoolMaxIdle),
		},
	}, nil
}

// Explode array of vulnerabilities type into a string separated by commas.
func GetVulnerabilities(vulnType []goharborv1alpha2.TrivyServerVulnerabilityType) string {
	vulnerabilities := ""

	for index, v := range vulnType {
		if index == 0 {
			vulnerabilities = string(v)
		} else {
			vulnerabilities = fmt.Sprintf("%s,%s", vulnerabilities, v)
		}
	}

	return vulnerabilities
}

// Explode array of severities type into a string separated by commas.
func GetSeverities(sevType []goharborv1alpha2.TrivyServerSeverityType) string {
	severities := ""

	for index, s := range sevType {
		if index == 0 {
			severities = string(s)
		} else {
			severities = fmt.Sprintf("%s,%s", severities, s)
		}
	}

	return severities
}
