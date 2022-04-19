package cache

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GenerateResourceList generates resource list by parsing parameters cpu and memory.
func GenerateResourceList(cpu, memory string) (resources corev1.ResourceList, err error) {
	resources = corev1.ResourceList{}
	if cpu != "" {
		resources[corev1.ResourceCPU], err = resource.ParseQuantity(cpu)
		if err != nil {
			return resources, err
		}
	}

	if memory != "" {
		resources[corev1.ResourceMemory], err = resource.ParseQuantity(memory)
		if err != nil {
			return resources, err
		}
	}

	return resources, nil
}

// GeneratePVC generates pvc by name and size.
func GenerateStoragePVC(storageClass, name, size string, labels map[string]string) (*corev1.PersistentVolumeClaim, error) {
	storage, err := resource.ParseQuantity(size)
	if err != nil {
		return nil, err
	}

	var sc *string
	if storageClass != "" {
		sc = &storageClass
	}

	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: sc,
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{"storage": storage},
			},
		},
	}, nil
}

// MergeLabels merges all labels together and returns a new label.
func MergeLabels(allLabels ...map[string]string) map[string]string {
	lb := make(map[string]string)

	for _, label := range allLabels {
		for k, v := range label {
			lb[k] = v
		}
	}

	return lb
}

// IsEqual check two object is equal.
func IsEqual(obj1, obj2 interface{}) bool {
	return equality.Semantic.DeepEqual(obj1, obj2)
}
