package v1alpha1

import (
	"sync"

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var harborlog = logf.Log.WithName("harbor-resource")

func (r *Harbor) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-containerregistry-ovhcloud-com-v1alpha1-harbor,mutating=true,failurePolicy=fail,groups=containerregistry.ovhcloud.com,resources=harbors,verbs=create;update,versions=v1alpha1,name=mharbor.kb.io

var _ webhook.Defaulter = &Harbor{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Harbor) Default() {
	harborlog.Info("default", "name", r.Name)

	err := r.DefaultImages()
	if err != nil {
		harborlog.Error(err, "default images", "version", r.Spec.HarborVersion)
	}

	if r.Spec.Components.JobService.WorkerCount == 0 {
		r.Spec.Components.JobService.WorkerCount = 3
	}
}

var registerOnce sync.Once

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Harbor) DefaultImages() error {
	registerOnce.Do(RegisterDefaultVersion)

	images, err := GetImages(r.Spec.HarborVersion)
	if err != nil {
		return errors.Wrap(err, "cannot get images value")
	}

	if r.Spec.Components.Core.Image == "" {
		r.Spec.Components.Core.Image = images.Core
	}

	if r.Spec.Components.Registry.Image == "" {
		r.Spec.Components.Registry.Image = images.Registry
	}

	if r.Spec.Components.RegistryCtl.Image == "" {
		r.Spec.Components.RegistryCtl.Image = images.RegistryCtl
	}

	if r.Spec.Components.Portal.Image == "" {
		r.Spec.Components.Portal.Image = images.Portal
	}

	if r.Spec.Components.JobService.Image == "" {
		r.Spec.Components.JobService.Image = images.JobService
	}

	if r.Spec.Components.ChartMuseum != nil {
		if r.Spec.Components.ChartMuseum.Image == "" {
			r.Spec.Components.ChartMuseum.Image = images.ChartMuseum
		}
	}

	if r.Spec.Components.Clair != nil {
		if r.Spec.Components.Clair.Image == "" {
			r.Spec.Components.Clair.Image = images.Clair
		}
	}

	if r.Spec.Components.Notary != nil {
		if r.Spec.Components.Notary.NotaryDBMigratorImage == "" {
			r.Spec.Components.Notary.NotaryDBMigratorImage = images.NotaryDBMigrator
		}

		if r.Spec.Components.Notary.Server.Image == "" {
			r.Spec.Components.Notary.Server.Image = images.NotaryServer
		}

		if r.Spec.Components.Notary.Signer.Image == "" {
			r.Spec.Components.Notary.Signer.Image = images.NotarySigner
		}
	}

	return nil
}
