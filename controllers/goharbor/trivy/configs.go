package trivy

import (
	"context"
	"strconv"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/graph"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Reconciler) AddConfigMap(ctx context.Context, trivy *goharborv1.Trivy) (graph.Resource, error) {
	// Forge the ConfigMap resource
	cm, err := r.GetConfigMap(ctx, trivy)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	// Add config map to reconciler controller
	cmResource, err := r.Controller.AddConfigMapToManage(ctx, cm)
	if err != nil {
		return nil, errors.Wrap(err, "add")
	}

	return cmResource, nil
}

// GetConfigMap get the config map linked to the trivy deployment.
func (r *Reconciler) GetConfigMap(ctx context.Context, trivy *goharborv1.Trivy) (*corev1.ConfigMap, error) {
	name := r.NormalizeName(ctx, trivy.GetName())
	namespace := trivy.GetNamespace()

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},

		Data: map[string]string{
			"SCANNER_LOG_LEVEL": string(trivy.Spec.Log.Level),

			"SCANNER_API_SERVER_READ_TIMEOUT":  trivy.Spec.Server.ReadTimeout.Duration.String(),
			"SCANNER_API_SERVER_WRITE_TIMEOUT": trivy.Spec.Server.WriteTimeout.Duration.String(),
			"SCANNER_API_SERVER_IDLE_TIMEOUT":  trivy.Spec.Server.IdleTimeout.Duration.String(),

			"SCANNER_TRIVY_CACHE_DIR":      "/home/scanner/.cache/trivy",
			"SCANNER_TRIVY_REPORTS_DIR":    "/home/scanner/.cache/reports",
			"SCANNER_TRIVY_DEBUG_MODE":     strconv.FormatBool(trivy.Spec.Server.DebugMode),
			"SCANNER_TRIVY_VULN_TYPE":      trivy.Spec.TrivyVulnerabilityTypes.GetValue(),
			"SCANNER_TRIVY_SEVERITY":       trivy.Spec.TrivySeverityTypes.GetValue(),
			"SCANNER_TRIVY_IGNORE_UNFIXED": strconv.FormatBool(trivy.Spec.Server.IgnoreUnfixed),
			"SCANNER_TRIVY_SKIP_UPDATE":    strconv.FormatBool(trivy.Spec.Update.Skip),
			"SCANNER_TRIVY_OFFLINE_SCAN":   strconv.FormatBool(trivy.Spec.OfflineScan),
			"SCANNER_TRIVY_INSECURE":       strconv.FormatBool(trivy.Spec.Server.Insecure),

			"SCANNER_STORE_REDIS_NAMESPACE":    trivy.Spec.Redis.Namespace,
			"SCANNER_STORE_REDIS_SCAN_JOB_TTL": trivy.Spec.Redis.Jobs.ScanTTL.Duration.String(),

			"SCANNER_JOB_QUEUE_REDIS_NAMESPACE":    trivy.Spec.Redis.Jobs.Namespace,
			"SCANNER_JOB_QUEUE_WORKER_CONCURRENCY": "1", // More may corrupt trivy.db file

			"SCANNER_REDIS_POOL_IDLE_TIMEOUT":       trivy.Spec.Redis.Pool.IdleTimeout.Duration.String(),
			"SCANNER_REDIS_POOL_CONNECTION_TIMEOUT": trivy.Spec.Redis.Pool.ConnectionTimeout.Duration.String(),
			"SCANNER_REDIS_POOL_READ_TIMEOUT":       trivy.Spec.Redis.Pool.ReadTimeout.Duration.String(),
			"SCANNER_REDIS_POOL_WRITE_TIMEOUT":      trivy.Spec.Redis.Pool.WriteTimeout.Duration.String(),
			"SCANNER_REDIS_POOL_MAX_ACTIVE":         strconv.Itoa(trivy.Spec.Redis.Pool.MaxActive),
			"SCANNER_REDIS_POOL_MAX_IDLE":           strconv.Itoa(trivy.Spec.Redis.Pool.MaxIdle),
		},
	}, nil
}
