package harbor

import (
	"context"

	goharborv1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

func (r *Reconciler) ChartMuseumStorage(ctx context.Context, harbor *goharborv1.Harbor) goharborv1.ChartMuseumChartStorageDriverSpec {
	if harbor.Spec.ImageChartStorage.S3 != nil {
		return goharborv1.ChartMuseumChartStorageDriverSpec{
			Amazon: harbor.Spec.ImageChartStorage.S3.ChartMuseum(),
		}
	}

	if harbor.Spec.ImageChartStorage.Swift != nil {
		return goharborv1.ChartMuseumChartStorageDriverSpec{
			OpenStack: harbor.Spec.ImageChartStorage.Swift.ChartMuseum(),
		}
	}

	if harbor.Spec.ImageChartStorage.Azure != nil {
		return goharborv1.ChartMuseumChartStorageDriverSpec{
			Azure: harbor.Spec.ImageChartStorage.Azure.ChartMuseum(),
		}
	}

	if harbor.Spec.ImageChartStorage.Gcs != nil {
		return goharborv1.ChartMuseumChartStorageDriverSpec{
			Gcs: harbor.Spec.ImageChartStorage.Gcs.ChartMuseum(),
		}
	}

	if harbor.Spec.ImageChartStorage.Oss != nil {
		return goharborv1.ChartMuseumChartStorageDriverSpec{
			Oss: harbor.Spec.ImageChartStorage.Oss.ChartMuseum(),
		}
	}

	prefix := ""
	pvc := &harbor.Spec.ImageChartStorage.FileSystem.RegistryPersistentVolume.PersistentVolumeClaimVolumeSource

	if harbor.Spec.ImageChartStorage.FileSystem.ChartPersistentVolume != nil {
		pvc = &harbor.Spec.ImageChartStorage.FileSystem.ChartPersistentVolume.PersistentVolumeClaimVolumeSource
		prefix = harbor.Spec.ImageChartStorage.FileSystem.ChartPersistentVolume.Prefix
	}

	return goharborv1.ChartMuseumChartStorageDriverSpec{
		FileSystem: &goharborv1.ChartMuseumChartStorageDriverFilesystemSpec{
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: pvc,
			},
			Prefix: prefix,
		},
	}
}

func (r *Reconciler) JobServiceScanDataExportsStorage(ctx context.Context, harbor *goharborv1.Harbor) goharborv1.JobServiceStorageVolumeSpec {
	if harbor.Spec.JobService.Storage != nil && harbor.Spec.JobService.Storage.ScanDataExportsPersistentVolume != nil {
		pvc := &harbor.Spec.JobService.Storage.ScanDataExportsPersistentVolume.PersistentVolumeClaimVolumeSource

		return goharborv1.JobServiceStorageVolumeSpec{
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: pvc,
			},
		}
	}

	return goharborv1.JobServiceStorageVolumeSpec{
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
}

func (r *Reconciler) TrivyReportsStorage(ctx context.Context, harbor *goharborv1.Harbor) goharborv1.TrivyStorageVolumeSpec {
	if harbor.Spec.Trivy.Storage.ReportsPersistentVolume != nil {
		pvc := &harbor.Spec.Trivy.Storage.ReportsPersistentVolume.PersistentVolumeClaimVolumeSource

		return goharborv1.TrivyStorageVolumeSpec{
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: pvc,
			},
		}
	}

	return goharborv1.TrivyStorageVolumeSpec{
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
}

func (r *Reconciler) TrivyCacheStorage(ctx context.Context, harbor *goharborv1.Harbor) goharborv1.TrivyStorageVolumeSpec {
	if harbor.Spec.Trivy.Storage.CachePersistentVolume != nil {
		pvc := &harbor.Spec.Trivy.Storage.CachePersistentVolume.PersistentVolumeClaimVolumeSource

		return goharborv1.TrivyStorageVolumeSpec{
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: pvc,
			},
		}
	}

	return goharborv1.TrivyStorageVolumeSpec{
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
}

func (r *Reconciler) RegistryStorage(ctx context.Context, harbor *goharborv1.Harbor) goharborv1.RegistryStorageDriverSpec {
	if harbor.Spec.ImageChartStorage.S3 != nil {
		return goharborv1.RegistryStorageDriverSpec{
			S3: harbor.Spec.ImageChartStorage.S3.Registry(),
		}
	}

	if harbor.Spec.ImageChartStorage.Swift != nil {
		return goharborv1.RegistryStorageDriverSpec{
			Swift: harbor.Spec.ImageChartStorage.Swift.Registry(),
		}
	}

	if harbor.Spec.ImageChartStorage.Azure != nil {
		return goharborv1.RegistryStorageDriverSpec{
			Azure: harbor.Spec.ImageChartStorage.Azure.Registry(),
		}
	}

	if harbor.Spec.ImageChartStorage.Oss != nil {
		return goharborv1.RegistryStorageDriverSpec{
			Oss: harbor.Spec.ImageChartStorage.Oss.Registry(),
		}
	}

	if harbor.Spec.ImageChartStorage.Gcs != nil {
		return goharborv1.RegistryStorageDriverSpec{
			Gcs: harbor.Spec.ImageChartStorage.Gcs.Registry(),
		}
	}

	return goharborv1.RegistryStorageDriverSpec{
		FileSystem: &goharborv1.RegistryStorageDriverFilesystemSpec{
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &harbor.Spec.ImageChartStorage.FileSystem.RegistryPersistentVolume.PersistentVolumeClaimVolumeSource,
			},
			MaxThreads: harbor.Spec.ImageChartStorage.FileSystem.RegistryPersistentVolume.MaxThreads,
			Prefix:     harbor.Spec.ImageChartStorage.FileSystem.RegistryPersistentVolume.Prefix,
		},
	}
}
