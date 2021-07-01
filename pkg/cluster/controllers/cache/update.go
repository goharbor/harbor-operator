package cache

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/cluster/controllers/common"
	"github.com/goharbor/harbor-operator/pkg/cluster/k8s"
	"github.com/goharbor/harbor-operator/pkg/cluster/lcm"
	"github.com/goharbor/harbor-operator/pkg/resources/checksum"
	redisOp "github.com/spotahome/redis-operator/api/redisfailover/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// RollingUpgrades reconcile will rolling upgrades Redis sentinel cluster if resource upscale.
// It does:
// - check resource
// - update RedisFailovers CR resource.
func (rc *RedisController) RollingUpgrades(ctx context.Context, cluster *goharborv1.HarborCluster, actualObj, expectObj runtime.Object) (*lcm.CRStatus, error) {
	crdClient := rc.DClient.DynamicClient(ctx, k8s.WithResource(redisFailoversGVR), k8s.WithNamespace(cluster.Namespace))

	if expectObj == nil || actualObj == nil {
		return cacheUnknownStatus(), nil
	}

	expectCR := expectObj.(*redisOp.RedisFailover)
	unstructuredActualCR := actualObj.(*unstructured.Unstructured)
	actualCR := &redisOp.RedisFailover{}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredActualCR.UnstructuredContent(), actualCR); err != nil {
		return cacheNotReadyStatus(ErrorDefaultUnstructuredConverter, err.Error()), err
	}

	if !common.Equals(ctx, rc.Scheme, cluster, actualObj.(checksum.Dependency)) {
		rc.Log.Info(
			"Update Redis resource",
			"namespace", cluster.Namespace, "name", cluster.Name,
		)

		expectCR.ObjectMeta.SetResourceVersion(actualCR.ObjectMeta.GetResourceVersion())

		data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&expectCR)
		if err != nil {
			return cacheUnknownStatus(), nil
		}

		_, err = crdClient.Update(&unstructured.Unstructured{Object: data}, metav1.UpdateOptions{})
		if err != nil {
			return cacheUnknownStatus(), err
		}

		return nil, nil
	}

	return cacheUnknownStatus(), nil
}

func (rc *RedisController) Update(ctx context.Context, cluster *goharborv1.HarborCluster, actualObj, expectObj runtime.Object) (*lcm.CRStatus, error) {
	crStatus, err := rc.RollingUpgrades(ctx, cluster, actualObj, expectObj)
	if err != nil {
		return crStatus, err
	}

	return crStatus, nil
}
