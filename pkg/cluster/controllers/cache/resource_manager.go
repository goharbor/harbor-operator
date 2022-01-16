package cache

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/common"
	"github.com/goharbor/harbor-operator/pkg/config"
	"github.com/goharbor/harbor-operator/pkg/resources/checksum"
	"github.com/ovh/configstore"
	redisOp "github.com/spotahome/redis-operator/api/redisfailover/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ResourceManager defines the common interface of resources.
type ResourceManager interface {
	ResourceGetter
	// With the specified cluster
	WithCluster(cluster *goharborv1.HarborCluster) ResourceManager
}

// ResourceGetter gets resources.
type ResourceGetter interface {
	GetCacheCR(ctx context.Context, harborcluster *goharborv1.HarborCluster) (runtime.Object, error)
	GetCacheCRName() string
	GetResources() corev1.ResourceRequirements
	GetSecretName() string
	GetSecret() *corev1.Secret
	GetServerReplica() int
	GetClusterServerReplica() int
	GetStorageSize() string
}

var _ ResourceManager = &redisResourceManager{}

type redisResourceManager struct {
	cluster     *goharborv1.HarborCluster
	configStore *configstore.Store
	logger      logr.Logger
	scheme      *runtime.Scheme
}

const (
	defaultResourceCPU     = "1"
	defaultResourceMemory  = "2Gi"
	defaultResourceReplica = 3
	defaultStorageSize     = "1Gi"
)

const (
	labelApp = "goharbor.io/harbor-cluster"
)

// NewResourceManager constructs a new cache resource manager.
func NewResourceManager(store *configstore.Store, logger logr.Logger, scheme *runtime.Scheme) ResourceManager {
	return &redisResourceManager{
		configStore: store,
		logger:      logger,
		scheme:      scheme,
	}
}

// WithCluster get resources based on the specified cluster spec.
func (rm *redisResourceManager) WithCluster(cluster *goharborv1.HarborCluster) ResourceManager {
	rm.cluster = cluster

	return rm
}

// GetCacheCR gets cache cr instance.
func (rm *redisResourceManager) GetCacheCR(ctx context.Context, harborcluster *goharborv1.HarborCluster) (runtime.Object, error) {
	resources := rm.GetResources()
	pvc, _ := GenerateStoragePVC(rm.GetStorageClass(), rm.cluster.Name, rm.GetStorageSize(), rm.GetLabels())

	image, err := rm.GetImage(ctx, harborcluster)
	if err != nil {
		return nil, err
	}

	rf := &redisOp.RedisFailover{
		TypeMeta: metav1.TypeMeta{
			Kind:       redisOp.RFKind,
			APIVersion: "databases.spotahome.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      rm.GetCacheCRName(),
			Namespace: rm.cluster.Namespace,
			Labels:    rm.GetLabels(),
		},
		Spec: redisOp.RedisFailoverSpec{
			Redis: redisOp.RedisSettings{
				Replicas:  int32(rm.GetServerReplica()),
				Resources: resources,
				Storage: redisOp.RedisStorage{
					KeepAfterDeletion: true,
					PersistentVolumeClaim: &redisOp.EmbeddedPersistentVolumeClaim{
						EmbeddedObjectMetadata: redisOp.EmbeddedObjectMetadata{
							Name:        pvc.Name,
							Labels:      pvc.ObjectMeta.Labels,
							Annotations: pvc.ObjectMeta.Annotations,
						},
						TypeMeta: pvc.TypeMeta,
						Spec:     pvc.Spec,
					},
				},
				Image:            image,
				ImagePullPolicy:  rm.getImagePullPolicy(ctx, harborcluster),
				ImagePullSecrets: rm.getImagePullSecrets(ctx, harborcluster),
			},
			Sentinel: redisOp.SentinelSettings{
				Replicas:         int32(rm.GetClusterServerReplica()),
				Resources:        resources,
				Image:            image,
				ImagePullPolicy:  rm.getImagePullPolicy(ctx, harborcluster),
				ImagePullSecrets: rm.getImagePullSecrets(ctx, harborcluster),
			},
			Auth: redisOp.AuthSettings{SecretPath: rm.GetSecretName()},
		},
	}

	dependencies := checksum.New(rm.scheme)
	dependencies.Add(ctx, harborcluster, true)
	dependencies.AddAnnotations(rf)

	return rf, nil
}

// GetCacheCRName gets cache cr name.
func (rm *redisResourceManager) GetCacheCRName() string {
	return fmt.Sprintf("%s-%s", rm.cluster.Name, "redis")
}

// GetSecretName gets secret name.
func (rm *redisResourceManager) GetSecretName() string {
	return rm.GetCacheCRName()
}

// GetSecret gets redis secret.
func (rm *redisResourceManager) GetSecret() *corev1.Secret {
	name := rm.GetSecretName()

	const SecretLen = 8
	passStr := common.RandomString(SecretLen, "a")

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: rm.cluster.Namespace,
			Labels:    rm.GetLabels(),
		},
		StringData: map[string]string{
			"redis-password": passStr,
			"password":       passStr,
		},
	}
}

// GetLabels gets labels merged from cluster labels.
func (rm *redisResourceManager) GetLabels() map[string]string {
	dynLabels := map[string]string{
		"app.kubernetes.io/name":     "cache",
		"app.kubernetes.io/instance": rm.cluster.Namespace,
		labelApp:                     rm.cluster.Name,
	}

	return MergeLabels(dynLabels, rm.cluster.Labels)
}

// GetResourceList gets redis resources.
func (rm *redisResourceManager) GetResources() corev1.ResourceRequirements {
	resources := corev1.ResourceRequirements{}

	if rm.cluster.Spec.Cache.Spec.RedisFailover.Server == nil {
		resourcesList, _ := GenerateResourceList(defaultResourceCPU, defaultResourceMemory)
		resources.Limits = resourcesList
		resources.Requests = resourcesList
	}

	return rm.cluster.Spec.Cache.Spec.RedisFailover.Server.Resources
}

// GetServerReplica gets deployment replica.
func (rm *redisResourceManager) GetServerReplica() int {
	if rm.cluster.Spec.Cache.Spec.RedisFailover.Server == nil || rm.cluster.Spec.Cache.Spec.RedisFailover.Server.Replicas == 0 {
		return defaultResourceReplica
	}

	return rm.cluster.Spec.Cache.Spec.RedisFailover.Server.Replicas
}

// GetClusterServerReplica gets deployment replica of sentinel mode.
func (rm *redisResourceManager) GetClusterServerReplica() int {
	if rm.cluster.Spec.Cache.Spec.RedisFailover.Sentinel == nil || rm.cluster.Spec.Cache.Spec.RedisFailover.Sentinel.Replicas == 0 {
		return defaultResourceReplica
	}

	return rm.cluster.Spec.Cache.Spec.RedisFailover.Sentinel.Replicas
}

// GetStorageSize gets storage size.
func (rm *redisResourceManager) GetStorageSize() string {
	if rm.cluster.Spec.Cache.Spec.RedisFailover.Server == nil || rm.cluster.Spec.Cache.Spec.RedisFailover.Server.Storage == "" {
		return defaultStorageSize
	}

	return rm.cluster.Spec.Cache.Spec.RedisFailover.Server.Storage
}

// GetStorageClass gets the storage class name.
func (rm *redisResourceManager) GetStorageClass() string {
	if rm.cluster.Spec.Cache.Spec.RedisFailover != nil && rm.cluster.Spec.Cache.Spec.RedisFailover.Server != nil {
		return rm.cluster.Spec.Cache.Spec.RedisFailover.Server.StorageClassName
	}

	return ""
}

func (rm *redisResourceManager) getImagePullPolicy(_ context.Context, harborcluster *goharborv1.HarborCluster) corev1.PullPolicy {
	if harborcluster.Spec.Cache.Spec.RedisFailover.ImagePullPolicy != nil {
		return *harborcluster.Spec.Cache.Spec.RedisFailover.ImagePullPolicy
	}

	if harborcluster.Spec.ImageSource != nil && harborcluster.Spec.ImageSource.ImagePullPolicy != nil {
		return *harborcluster.Spec.ImageSource.ImagePullPolicy
	}

	return config.DefaultImagePullPolicy
}

func (rm *redisResourceManager) getImagePullSecrets(_ context.Context, harborcluster *goharborv1.HarborCluster) []corev1.LocalObjectReference {
	if len(harborcluster.Spec.Cache.Spec.RedisFailover.ImagePullSecrets) > 0 {
		return harborcluster.Spec.Cache.Spec.RedisFailover.ImagePullSecrets
	}

	if harborcluster.Spec.ImageSource != nil {
		return harborcluster.Spec.ImageSource.ImagePullSecrets
	}

	return nil
}
