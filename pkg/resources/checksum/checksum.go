package checksum

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/goharbor/harbor-operator/pkg/factories/logger"
	"github.com/goharbor/harbor-operator/pkg/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	versionAnnotationChecksumKey = "harbor.checksum.goharbor.io/version"
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

func (d *Dependencies) Add(ctx context.Context, resource Dependency, onlySpec bool) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.objects[resource] = onlySpec
}

func (d *Dependencies) GetID(resource Dependency) string {
	namespace := resource.GetNamespace()
	if namespace == "" {
		namespace = "unknown"
	}

	gvks, _, err := d.scheme.ObjectKinds(resource)
	if err != nil {
		return fmt.Sprintf("%s.unknown.checksum.goharbor.io/%s", namespace, resource.GetName())
	}

	return fmt.Sprintf("%s.%s.checksum.goharbor.io/%s", namespace, strings.ToLower(gvks[0].Kind), resource.GetName())
}

func GetStaticID(name string) string {
	return fmt.Sprintf("static.checksum.goharbor.io/%s", name)
}

func (d *Dependencies) ComputeChecksum(resource metav1.Object, onlySpec bool) string {
	if !onlySpec {
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

	for object, onlySpec := range d.objects {
		previous, ok := annotations[d.GetID(object)]
		if !ok {
			logger.Get(ctx).V(1).Info("dependencies changed (no annotation)", "dependency.kind", object.GetObjectKind(), "dependency", object)

			return true
		}

		current := d.ComputeChecksum(object, onlySpec)
		if previous != current {
			logger.Get(ctx).V(1).Info(fmt.Sprintf("dependencies changed (expected %s, got %s)", previous, current), "dependency.kind", object.GetObjectKind(), "dependency", object)

			return true
		}
	}

	if current := version.GetVersion(annotations); current != "" {
		previous, ok := annotations[versionAnnotationChecksumKey]
		if !ok {
			logger.Get(ctx).V(1).Info("version changed (no annotation)")

			return true
		}

		if previous != current {
			logger.Get(ctx).V(1).Info("version changed (expected %s, got %s)", previous, current)

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

	for obj, onlySpec := range d.objects {
		annotations[d.GetID(obj)] = d.ComputeChecksum(obj, onlySpec)
	}

	if ver := version.GetVersion(annotations); ver != "" {
		annotations[versionAnnotationChecksumKey] = ver
	}

	object.SetAnnotations(annotations)
}

func CopyVersion(from, to metav1.Object) {
	to.SetUID(from.GetUID())
	to.SetGeneration(from.GetGeneration())
	to.SetResourceVersion(from.GetResourceVersion())
}

func CopyMarkers(from, to metav1.Object) {
	toAnnotations := to.GetAnnotations()
	if toAnnotations == nil {
		toAnnotations = map[string]string{}
	}

	fromAnnotations := from.GetAnnotations()
	if fromAnnotations == nil {
		fromAnnotations = map[string]string{}
	}

	for key := range toAnnotations {
		if !strings.Contains(key, ".checksum.goharbor.io/") {
			continue
		}

		if IsStaticAnnotation(key) {
			continue
		}

		delete(toAnnotations, key)
	}

	for key, value := range fromAnnotations {
		if !strings.Contains(key, ".checksum.goharbor.io/") {
			continue
		}

		if IsStaticAnnotation(key) {
			continue
		}

		toAnnotations[key] = value
	}

	to.SetAnnotations(toAnnotations)
}

func IsStaticAnnotation(key string) bool {
	return strings.HasPrefix(key, "static.checksum.goharbor.io/")
}
