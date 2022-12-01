package v1beta1

import (
	"context"
	"fmt"

	"github.com/goharbor/harbor-operator/pkg/version"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (harborcluster *HarborCluster) SetupWebhookWithManager(_ context.Context, mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(harborcluster).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

const (
	// The minimal number of volumes created by minio.
	minimalVolumeCount = 4
)

// Log used this webhook.
var clog = logf.Log.WithName("harborcluster-resource")

// +kubebuilder:webhook:verbs=create;update,path=/mutate-goharbor-io-v1beta1-harborcluster,mutating=true,failurePolicy=fail,groups=goharbor.io,resources=harborclusters,versions=v1beta1,name=mharborcluster.kb.io,admissionReviewVersions={"v1beta1","v1"},sideEffects=None

var _ webhook.Defaulter = &HarborCluster{}

func (harborcluster *HarborCluster) Default() { //nolint:funlen
	switch harborcluster.Spec.Cache.Kind {
	case KindCacheRedis:
		harborcluster.Spec.Cache.Spec.RedisFailover = nil
	case KindCacheRedisFailover:
		harborcluster.Spec.Cache.Spec.Redis = nil
	}

	switch harborcluster.Spec.Database.Kind {
	case KindDatabasePostgreSQL:
		harborcluster.Spec.Database.Spec.ZlandoPostgreSQL = nil
	case KindDatabaseZlandoPostgreSQL:
		harborcluster.Spec.Database.Spec.PostgreSQL = nil
	}

	switch harborcluster.Spec.Storage.Kind {
	case KindStorageFileSystem:
		harborcluster.Spec.Storage.Spec.Oss = nil
		harborcluster.Spec.Storage.Spec.Gcs = nil
		harborcluster.Spec.Storage.Spec.Azure = nil
		harborcluster.Spec.Storage.Spec.S3 = nil
		harborcluster.Spec.Storage.Spec.Swift = nil
		harborcluster.Spec.Storage.Spec.MinIO = nil
	case KindStorageS3:
		harborcluster.Spec.Storage.Spec.Oss = nil
		harborcluster.Spec.Storage.Spec.Gcs = nil
		harborcluster.Spec.Storage.Spec.Azure = nil
		harborcluster.Spec.Storage.Spec.FileSystem = nil
		harborcluster.Spec.Storage.Spec.Swift = nil
		harborcluster.Spec.Storage.Spec.MinIO = nil
	case KindStorageSwift:
		harborcluster.Spec.Storage.Spec.Oss = nil
		harborcluster.Spec.Storage.Spec.Gcs = nil
		harborcluster.Spec.Storage.Spec.Azure = nil
		harborcluster.Spec.Storage.Spec.S3 = nil
		harborcluster.Spec.Storage.Spec.FileSystem = nil
		harborcluster.Spec.Storage.Spec.MinIO = nil
	case KindStorageMinIO:
		harborcluster.Spec.Storage.Spec.Oss = nil
		harborcluster.Spec.Storage.Spec.Gcs = nil
		harborcluster.Spec.Storage.Spec.Azure = nil
		harborcluster.Spec.Storage.Spec.S3 = nil
		harborcluster.Spec.Storage.Spec.Swift = nil
		harborcluster.Spec.Storage.Spec.FileSystem = nil
	case KindStorageAzure:
		harborcluster.Spec.Storage.Spec.Oss = nil
		harborcluster.Spec.Storage.Spec.Gcs = nil
		harborcluster.Spec.Storage.Spec.S3 = nil
		harborcluster.Spec.Storage.Spec.Swift = nil
		harborcluster.Spec.Storage.Spec.FileSystem = nil
		harborcluster.Spec.Storage.Spec.MinIO = nil
	case KindStorageGcs:
		harborcluster.Spec.Storage.Spec.Oss = nil
		harborcluster.Spec.Storage.Spec.Azure = nil
		harborcluster.Spec.Storage.Spec.S3 = nil
		harborcluster.Spec.Storage.Spec.Swift = nil
		harborcluster.Spec.Storage.Spec.FileSystem = nil
		harborcluster.Spec.Storage.Spec.MinIO = nil
	case KindStorageOss:
		harborcluster.Spec.Storage.Spec.Azure = nil
		harborcluster.Spec.Storage.Spec.S3 = nil
		harborcluster.Spec.Storage.Spec.Swift = nil
		harborcluster.Spec.Storage.Spec.FileSystem = nil
		harborcluster.Spec.Storage.Spec.MinIO = nil
		harborcluster.Spec.Storage.Spec.Gcs = nil
	}
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-goharbor-io-v1beta1-harborcluster,mutating=false,failurePolicy=fail,groups=goharbor.io,resources=harborclusters,versions=v1beta1,name=vharborcluster.kb.io,admissionReviewVersions={"v1beta1","v1"},sideEffects=None

var _ webhook.Validator = &HarborCluster{}

func (harborcluster *HarborCluster) ValidateCreate() error {
	clog.Info("validate creation", "name", harborcluster.Name, "namespace", harborcluster.Namespace)

	return harborcluster.validate(harborcluster)
}

func (harborcluster *HarborCluster) ValidateUpdate(old runtime.Object) error {
	clog.Info("validate updating", "name", harborcluster.Name, "namespace", harborcluster.Namespace)

	obj, ok := old.(*HarborCluster)
	if !ok {
		return errors.Errorf("failed type assertion on kind: %s", old.GetObjectKind().GroupVersionKind().String())
	}

	return harborcluster.validate(obj)
}

func (harborcluster *HarborCluster) ValidateDelete() error {
	clog.Info("validate deletion", "name", harborcluster.Name, "namespace", harborcluster.Namespace)

	return nil
}

func (harborcluster *HarborCluster) validate(old *HarborCluster) error {
	var allErrs field.ErrorList

	if err := harborcluster.Spec.Network.Validate(nil); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := harborcluster.Spec.Trace.Validate(nil); err != nil {
		allErrs = append(allErrs, err)
	}

	// For database(psql), cache(Redis) and storage, either external services or in-cluster services MUST be configured
	if err := harborcluster.validateStorage(); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := harborcluster.validateDatabase(); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := harborcluster.validateCache(); err != nil {
		allErrs = append(allErrs, err)
	}

	if old == nil { // create harbor resource
		if err := version.Validate(harborcluster.Spec.Version); err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("version"), harborcluster.Spec.Version, err.Error()))
		}
	} else {
		if err := version.UpgradeAllowed(old.Spec.Version, harborcluster.Spec.Version); err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("version"), harborcluster.Spec.Version, err.Error()))
		}
	}

	if old.Spec.Cache.Kind != harborcluster.Spec.Cache.Kind {
		allErrs = append(allErrs, field.Forbidden(
			field.NewPath("spec").Child("cache"),
			"don't allow to switch cache between incluster and external"))
	}

	if old.Spec.Database.Kind != harborcluster.Spec.Database.Kind {
		allErrs = append(allErrs, field.Forbidden(
			field.NewPath("spec").Child("database"),
			"don't allow to switch database between incluster and external"))
	}

	if old.Spec.Storage.Kind != harborcluster.Spec.Storage.Kind {
		allErrs = append(allErrs, field.Forbidden(
			field.NewPath("spec").Child("storage"),
			"don't allow to switch storage between incluster and external"))
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "HarborCluster"}, harborcluster.Name, allErrs)
}

func (harborcluster *HarborCluster) validateStorage() *field.Error { //nolint:funlen,gocognit
	// in cluster storage has high priority
	fp := field.NewPath("spec").Child("storage").Child("spec")

	// Storage
	if harborcluster.Spec.Storage.Kind == KindStorageS3 && harborcluster.Spec.Storage.Spec.S3 == nil {
		// Invalid and not acceptable
		return required(fp.Child("s3"))
	}

	if harborcluster.Spec.Storage.Kind == KindStorageMinIO && harborcluster.Spec.Storage.Spec.MinIO == nil {
		// Invalid and not acceptable
		return required(fp.Child("minIO"))
	}

	if harborcluster.Spec.Storage.Kind == KindStorageSwift && harborcluster.Spec.Storage.Spec.Swift == nil {
		// Invalid and not acceptable
		return required(fp.Child("swift"))
	}

	if harborcluster.Spec.Storage.Kind == KindStorageAzure && harborcluster.Spec.Storage.Spec.Azure == nil {
		// Invalid and not acceptable
		return required(fp.Child("azure"))
	}

	if harborcluster.Spec.Storage.Kind == KindStorageGcs && harborcluster.Spec.Storage.Spec.Gcs == nil {
		// Invalid and not acceptable
		return required(fp.Child("gcs"))
	}

	if harborcluster.Spec.Storage.Kind == KindStorageOss && harborcluster.Spec.Storage.Spec.Oss == nil {
		// Invalid and not acceptable
		return required(fp.Child("oss"))
	}

	if harborcluster.Spec.Storage.Kind == KindStorageFileSystem && harborcluster.Spec.Storage.Spec.FileSystem == nil {
		// Invalid and not acceptable
		return required(fp.Child("fileSystem"))
	}

	// Validate more if incluster storage is configured.
	if harborcluster.Spec.Storage.Kind == KindStorageMinIO { //nolint:nestif
		desiredReplicas := harborcluster.Spec.Storage.Spec.MinIO.Replicas
		volumePerServer := harborcluster.Spec.Storage.Spec.MinIO.VolumesPerServer

		if desiredReplicas*volumePerServer < minimalVolumeCount {
			return invalid(fp, harborcluster.Spec.Storage.Spec.MinIO, fmt.Sprintf("minIO.replicas * minIO.volumesPerServer should be >=%d", minimalVolumeCount))
		}

		redirect := harborcluster.Spec.Storage.Spec.Redirect
		rp := fp.Child("redirect")

		if redirect == nil && harborcluster.Spec.Storage.Spec.MinIO != nil {
			redirect = harborcluster.Spec.Storage.Spec.MinIO.Redirect
			rp = fp.Child("minio").Child("redirect")
		}

		if redirect != nil && redirect.Enable {
			if redirect.Expose == nil || redirect.Expose.Ingress == nil {
				return required(rp.Child("ingress"))
			}
		}
	}

	return nil
}

func (harborcluster *HarborCluster) validateDatabase() *field.Error {
	// in cluster database has high priority
	fp := field.NewPath("spec").Child("database").Child("spec")

	// Database
	// External is not configured
	// And also in-cluster psql is not specified
	if harborcluster.Spec.Database.Kind == KindDatabasePostgreSQL && harborcluster.Spec.Database.Spec.PostgreSQL == nil {
		// Invalid and not acceptable
		return required(fp.Child("postgreSQL"))
	}

	// Both are configured then conflict
	if harborcluster.Spec.Database.Kind == KindDatabaseZlandoPostgreSQL && harborcluster.Spec.Database.Spec.ZlandoPostgreSQL == nil {
		// Invalid and not acceptable
		return required(fp.Child("zlandoPostgreSQL"))
	}

	return nil
}

func (harborcluster *HarborCluster) validateCache() *field.Error {
	fp := field.NewPath("spec").Child("cache").Child("spec")

	// Cache
	// External is not configured
	if harborcluster.Spec.Cache.Kind == KindCacheRedis && harborcluster.Spec.Cache.Spec.Redis == nil {
		// Invalid and not acceptable
		return required(fp.Child("redis"))
	}

	// Both are configured and then conflict
	if harborcluster.Spec.Cache.Kind == KindCacheRedisFailover && harborcluster.Spec.Cache.Spec.RedisFailover == nil {
		// Invalid and not acceptable
		return required(fp.Child("redisFailover"))
	}

	return nil
}

func required(mainPath *field.Path) *field.Error {
	return field.Required(mainPath, fmt.Sprintf("%s should be configured", mainPath.String()))
}

func invalid(mainPath *field.Path, value interface{}, details string) *field.Error {
	return field.Invalid(mainPath, value, details)
}
