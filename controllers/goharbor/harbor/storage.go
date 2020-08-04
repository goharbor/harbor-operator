package harbor

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
)

func (r *Reconciler) ChartMuseumStorage(ctx context.Context, harbor *goharborv1alpha2.Harbor) goharborv1alpha2.ChartMuseumChartStorageDriverSpec {
	if harbor.Spec.ImageChartStorage.S3 != nil {
		return goharborv1alpha2.ChartMuseumChartStorageDriverSpec{
			Amazon: harbor.Spec.ImageChartStorage.S3.ChartMuseum(),
		}
	}

	if harbor.Spec.ImageChartStorage.Swift != nil {
		return goharborv1alpha2.ChartMuseumChartStorageDriverSpec{
			OpenStack: harbor.Spec.ImageChartStorage.Swift.ChartMuseum(),
		}
	}

	prefix := ""
	pvc := &harbor.Spec.ImageChartStorage.FileSystem.RegistryPersistentVolume.PersistentVolumeClaimVolumeSource

	if harbor.Spec.ImageChartStorage.FileSystem.ChartPersistentVolume != nil {
		pvc = &harbor.Spec.ImageChartStorage.FileSystem.ChartPersistentVolume.PersistentVolumeClaimVolumeSource
		prefix = harbor.Spec.ImageChartStorage.FileSystem.ChartPersistentVolume.Prefix
	}

	return goharborv1alpha2.ChartMuseumChartStorageDriverSpec{
		FileSystem: &goharborv1alpha2.ChartMuseumChartStorageDriverFilesystemSpec{
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: pvc,
			},
			Prefix: prefix,
		},
	}
}

func (r *Reconciler) RegistryStorage(ctx context.Context, harbor *goharborv1alpha2.Harbor) goharborv1alpha2.RegistryStorageDriverSpec {
	if harbor.Spec.ImageChartStorage.S3 != nil {
		return goharborv1alpha2.RegistryStorageDriverSpec{
			S3: harbor.Spec.ImageChartStorage.S3.Registry(),
		}
	}

	if harbor.Spec.ImageChartStorage.Swift != nil {
		return goharborv1alpha2.RegistryStorageDriverSpec{
			Swift: harbor.Spec.ImageChartStorage.Swift.Registry(),
		}
	}

	return goharborv1alpha2.RegistryStorageDriverSpec{
		FileSystem: &goharborv1alpha2.RegistryStorageDriverFilesystemSpec{
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &harbor.Spec.ImageChartStorage.FileSystem.RegistryPersistentVolume.PersistentVolumeClaimVolumeSource,
			},
			MaxThreads: harbor.Spec.ImageChartStorage.FileSystem.RegistryPersistentVolume.MaxThreads,
			Prefix:     harbor.Spec.ImageChartStorage.FileSystem.RegistryPersistentVolume.Prefix,
		},
	}
}
