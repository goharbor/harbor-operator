package exporter

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/controllers"
	serrors "github.com/goharbor/harbor-operator/pkg/controller/errors"
	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/image"
	"github.com/goharbor/harbor-operator/pkg/version"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ConfigPath                            = "/etc/exporter"
	HealthPath                            = "/"
	InternalCertificatesVolumeName        = "internal-certificates"
	InternalCertificateAuthorityDirectory = "/harbor_cust_cert"
	InternalCertificatesPath              = ConfigPath + "/ssl"
)

var (
	varFalse = false

	fsGroup    int64 = 10000
	runAsGroup int64 = 10000
	runAsUser  int64 = 10000

	metricNamespace = "harbor"
	metricSubsytem  = "exporter"

	jobserivceRedisPasswordKey = "redis-password"
	jobserivceRedisTimeout     = 3600
	jobserviceRedisNamespace   = "harbor_job_service_namespace"
)

func (r *Reconciler) GetDeployment(ctx context.Context, exporter *goharborv1.Exporter) (*appsv1.Deployment, error) { //nolint:funlen
	getImageOptions := []image.Option{
		image.WithImageFromSpec(exporter.Spec.Image),
		image.WithHarborVersion(version.GetVersion(exporter.Annotations)),
	}

	image, err := image.GetImage(ctx, harbormetav1.ExporterComponent.String(), getImageOptions...)
	if err != nil {
		return nil, errors.Wrap(err, "get image")
	}

	name := r.NormalizeName(ctx, exporter.GetName())
	namespace := exporter.GetNamespace()

	serviceScheme, serviceHost, servicePort, err := parseServiceInfo(exporter.Spec.Core.URL)
	if err != nil {
		return nil, serrors.UnrecoverrableError(err, serrors.InvalidSpecReason, "invalid core url")
	}

	// Only one host is supported
	if len(exporter.Spec.Database.Hosts) == 0 {
		return nil, serrors.UnrecoverrableError(harbormetav1.NewErrPostgresNoHost(), serrors.InvalidSpecReason, "get a database host")
	}

	dbHost := exporter.Spec.Database.Hosts[0]

	envs := []corev1.EnvVar{
		{Name: "LOG_LEVEL", Value: exporter.Spec.Log.Level.String()},
		{Name: "HARBOR_EXPORTER_PORT", Value: strconv.Itoa(int(exporter.Spec.Port))},
		{Name: "HARBOR_EXPORTER_METRICS_PATH", Value: exporter.Spec.Path},
		{Name: "HARBOR_EXPORTER_METRICS_ENABLED", Value: "true"},
		{Name: "HARBOR_EXPORTER_CACHE_TIME", Value: exporter.Spec.Cache.GetDurationEnvVar()},
		{Name: "HARBOR_EXPORTER_CACHE_CLEAN_INTERVAL", Value: exporter.Spec.Cache.GetCleanIntervalEnvVar()},
		{Name: "HARBOR_METRIC_NAMESPACE", Value: metricNamespace},
		{Name: "HARBOR_METRIC_SUBSYSTEM", Value: metricSubsytem},
		{Name: "HARBOR_SERVICE_SCHEME", Value: serviceScheme},
		{Name: "HARBOR_SERVICE_HOST", Value: serviceHost},
		{Name: "HARBOR_SERVICE_PORT", Value: servicePort},
		{Name: "HARBOR_DATABASE_HOST", Value: dbHost.Host},
		{Name: "HARBOR_DATABASE_PORT", Value: strconv.Itoa(int(dbHost.Port))},
		{Name: "HARBOR_DATABASE_USERNAME", Value: exporter.Spec.Database.Username},
		{Name: "HARBOR_DATABASE_PASSWORD", ValueFrom: exporter.Spec.Database.GetPasswordEnvVarSource()},
		{Name: "HARBOR_DATABASE_DBNAME", Value: exporter.Spec.Database.Database},
	}

	if sslMode, ok := exporter.Spec.Database.Parameters[harbormetav1.PostgresSSLModeKey]; ok {
		envs = append(envs, corev1.EnvVar{
			Name:  "HARBOR_DATABASE_SSLMODE",
			Value: sslMode,
		})
	}

	if exporter.Spec.Database.MaxIdleConnections != nil {
		envs = append(envs, corev1.EnvVar{
			Name:  "HARBOR_DATABASE_MAX_IDLE_CONNS",
			Value: strconv.Itoa(int(*exporter.Spec.Database.MaxIdleConnections)),
		})
	}

	if exporter.Spec.Database.MaxOpenConnections != nil {
		envs = append(envs, corev1.EnvVar{
			Name:  "HARBOR_DATABASE_MAX_OPEN_CONNS",
			Value: strconv.Itoa(int(*exporter.Spec.Database.MaxOpenConnections)),
		})
	}

	redisURL, err := r.getJobServiceRedisURL(ctx, exporter)
	if err != nil {
		return nil, errors.Wrap(err, "get redis url of jobservice")
	}

	redisNamespace := jobserviceRedisNamespace
	if exporter.Spec.JobService.Redis.Namespace != "" {
		redisNamespace = exporter.Spec.JobService.Redis.Namespace
	}

	redisTimeout := jobserivceRedisTimeout
	if exporter.Spec.JobService.Redis.IdleTimeout != nil {
		redisTimeout = int(exporter.Spec.JobService.Redis.IdleTimeout.Seconds())
	}

	envs = append(envs, []corev1.EnvVar{
		{Name: "HARBOR_REDIS_URL", Value: redisURL},
		{Name: "HARBOR_REDIS_NAMESPACE", Value: redisNamespace},
		{Name: "HARBOR_REDIS_TIMEOUT", Value: fmt.Sprintf("%d", redisTimeout)},
	}...)

	volumes := []corev1.Volume{}

	volumeMounts := []corev1.VolumeMount{}

	if exporter.Spec.TLS.Enabled() {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      InternalCertificatesVolumeName,
			MountPath: path.Join(InternalCertificateAuthorityDirectory, corev1.ServiceAccountRootCAKey),
			SubPath:   strings.TrimLeft(corev1.ServiceAccountRootCAKey, "/"),
			ReadOnly:  true,
		}, corev1.VolumeMount{
			Name:      InternalCertificatesVolumeName,
			MountPath: InternalCertificatesPath,
			ReadOnly:  true,
		})

		volumes = append(volumes, corev1.Volume{
			Name: InternalCertificatesVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: exporter.Spec.TLS.CertificateRef,
				},
			},
		})
	} else {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      InternalCertificatesVolumeName,
			MountPath: InternalCertificateAuthorityDirectory,
		})

		volumes = append(volumes, corev1.Volume{
			Name: InternalCertificatesVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	httpGET := &corev1.HTTPGetAction{
		Path:   HealthPath,
		Port:   intstr.FromString(harbormetav1.ExporterMetricsPortName),
		Scheme: corev1.URISchemeHTTP,
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: version.NewVersionAnnotations(exporter.Annotations),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					r.Label("name"):      name,
					r.Label("namespace"): namespace,
				},
			},
			Replicas: exporter.Spec.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: exporter.Spec.ComponentSpec.TemplateAnnotations,
					Labels: map[string]string{
						r.Label("name"):      name,
						r.Label("namespace"): namespace,
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector:                 exporter.Spec.NodeSelector,
					AutomountServiceAccountToken: &varFalse,
					Volumes:                      volumes,
					SecurityContext: &corev1.PodSecurityContext{
						FSGroup:    &fsGroup,
						RunAsGroup: &runAsGroup,
						RunAsUser:  &runAsUser,
					},
					Containers: []corev1.Container{{
						Name:  controllers.Exporter.String(),
						Image: image,
						Args: []string{
							"-log-level",
							exporter.Spec.Log.Level.String(),
						},
						Ports: []corev1.ContainerPort{{
							Name:          harbormetav1.ExporterMetricsPortName,
							ContainerPort: exporter.Spec.Port,
							Protocol:      corev1.ProtocolTCP,
						}},

						Env: envs,

						VolumeMounts: volumeMounts,

						LivenessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: httpGET,
							},
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: httpGET,
							},
						},
					}},
				},
			},
		},
	}

	exporter.Spec.ComponentSpec.ApplyToDeployment(deploy)

	return deploy, nil
}

func (r *Reconciler) getJobServiceRedisURL(ctx context.Context, exporter *goharborv1.Exporter) (string, error) {
	if exporter.Spec.JobService == nil {
		// compatible with the exporter converted from v1alpha3 and the upgrade to harbor v2.3.x
		key := client.ObjectKey{
			Namespace: exporter.Namespace,
			Name:      strings.TrimSuffix(exporter.Name, "-harbor"),
		}

		var harbor goharborv1.Harbor
		if err := r.Client.Get(ctx, key, &harbor); err != nil {
			return "", err
		}

		exporter.Spec.JobService = &goharborv1.ExporterJobServiceSpec{
			Redis: &goharborv1.JobServicePoolRedisSpec{
				RedisConnection: harbor.Spec.RedisConnection(harbormetav1.JobServiceRedis),
			},
		}
	}

	redisPassword, err := r.getValueFromSecret(ctx, exporter.GetNamespace(), exporter.Spec.JobService.Redis.PasswordRef, jobserivceRedisPasswordKey)
	if err != nil {
		return "", errors.Wrap(err, "get redis password of jobservice")
	}

	if redisPassword == "" {
		logger.Get(ctx).Info("redis password secret of jobservice not found", "secret", exporter.Spec.JobService.Redis.PasswordRef)
	}

	return exporter.Spec.JobService.Redis.GetDSNStringWithRawPassword(redisPassword), nil
}

func (r *Reconciler) getValueFromSecret(ctx context.Context, namespace, name, key string) (string, error) {
	var secret corev1.Secret

	err := r.Client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &secret)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return "", nil
		}

		return "", err
	}

	for k, d := range secret.Data {
		if k == key {
			return string(d), nil
		}
	}

	return "", nil
}

func parseServiceInfo(coreURL string) (scheme string, host string, port string, err error) {
	u, err := url.Parse(coreURL)
	if err != nil {
		return "", "", "", err
	}

	scheme = strings.ToLower(u.Scheme)
	host = u.Hostname()
	port = u.Port()

	if port == "" {
		switch scheme {
		case "http":
			port = "80"
		default:
			port = "443"
		}
	}

	return
}
