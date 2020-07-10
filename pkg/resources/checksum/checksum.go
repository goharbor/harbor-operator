package checksum

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Dependencies struct {
	scheme  *runtime.Scheme
	objects map[Dependency]bool
	lock    sync.RWMutex
}

type Dependency interface {
	runtime.Object
	metav1.Object
}

func New(scheme *runtime.Scheme) *Dependencies {
	return &Dependencies{
		scheme:  scheme,
		objects: map[Dependency]bool{},
	}
}

func (d *Dependencies) Add(ctx context.Context, resource Dependency, withStatus bool) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.objects[resource] = withStatus
}

func (d *Dependencies) GetID(resource Dependency) string {
	gvks, _, err := d.scheme.ObjectKinds(resource)
	if err != nil {
		return fmt.Sprintf("%s.unknown.checksum.goharbor.io/%s", resource.GetNamespace(), resource.GetName())
	}

	return fmt.Sprintf("%s.%s.checksum.goharbor.io/%s", resource.GetNamespace(), strings.ToLower(gvks[0].Kind), resource.GetName())
}

func (d *Dependencies) ComputeChecksum(resource metav1.Object, withStatus bool) string {
	if withStatus {
		return resource.GetResourceVersion()
	}

	return fmt.Sprintf("%d", resource.GetGeneration())
}

func (d *Dependencies) ChangedFor(ctx context.Context, resource Dependency) bool {
	d.lock.RLock()
	defer d.lock.RUnlock()

	annotations := resource.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	for object, withStatus := range d.objects {
		previous, ok := annotations[d.GetID(object)]
		if !ok {
			gvks, _, err := d.scheme.ObjectKinds(resource)
			if err != nil {
				return true
			}
			logger.Get(ctx).V(1).Info("dependencies changed (no annotation)", "resource", resource, "dependency", object, "annotations", annotations, "gvk", gvks[0])
			return true
		}

		current := d.ComputeChecksum(object, withStatus)
		if previous != current {
			logger.Get(ctx).V(1).Info(fmt.Sprintf("dependencies changed (expected %s, got %s)", previous, current), "resource", resource, "dependency", object)
			return true
		}
	}

	return false
}

func (d *Dependencies) AddAnnotations(object metav1.Object) {
	annotations := object.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	for object, withStatus := range d.objects {
		annotations[d.GetID(object)] = d.ComputeChecksum(object, withStatus)
	}

	object.SetAnnotations(annotations)
}

func CopyMarkers(from, to metav1.Object) {
	to.SetUID(from.GetUID())
	to.SetGeneration(from.GetGeneration())
	to.SetResourceVersion(from.GetResourceVersion())

	annotations := to.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	for key, value := range from.GetAnnotations() {
		if strings.Contains(key, ".checksum.goharbor.io/") {
			annotations[key] = value
		}
	}

	to.SetAnnotations(annotations)
}
