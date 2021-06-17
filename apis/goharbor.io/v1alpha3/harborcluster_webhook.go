// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha3

import (
	"context"
	"fmt"

	"github.com/goharbor/harbor-operator/pkg/version"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const (
	// The minimal number of volumes created by minio.
	minimalVolumeCount = 4
)

// Log used this webhook.
var clog = logf.Log.WithName("harborcluster-resource")

// SetupWebhookWithManager sets up validating webhook of HarborCluster.
func (hc *HarborCluster) SetupWebhookWithManager(_ context.Context, mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(hc).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-goharbor-io-v1alpha3-harborcluster,mutating=true,failurePolicy=fail,groups=goharbor.io,resources=harborclusters,verbs=create;update,versions=v1alpha3,name=mharborcluster.kb.io,admissionReviewVersions={"v1beta1"},sideEffects=None

var _ webhook.Defaulter = &HarborCluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (hc *HarborCluster) Default() {
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-goharbor-io-v1alpha3-harborcluster,mutating=false,failurePolicy=fail,groups=goharbor.io,resources=harborclusters,versions=v1alpha3,name=vharborcluster.kb.io,admissionReviewVersions={"v1beta1"},sideEffects=None

var _ webhook.Validator = &HarborCluster{}

func (hc *HarborCluster) ValidateCreate() error {
	clog.Info("validate creation", "name", hc.Name, "namespace", hc.Namespace)

	return hc.validate(hc)
}

func (hc *HarborCluster) ValidateUpdate(old runtime.Object) error {
	clog.Info("validate updating", "name", hc.Name, "namespace", hc.Namespace)

	obj, ok := old.(*HarborCluster)
	if !ok {
		return errors.Errorf("failed type assertion on kind: %s", old.GetObjectKind().GroupVersionKind().String())
	}

	return hc.validate(obj)
}

func (hc *HarborCluster) ValidateDelete() error {
	clog.Info("validate deletion", "name", hc.Name, "namespace", hc.Namespace)

	return nil
}

func (hc *HarborCluster) validate(old *HarborCluster) error {
	var allErrs field.ErrorList

	clog.Info("harbor cluster", "value", hc)

	// For database(psql), cache(Redis) and storage, either external services or in-cluster services MUST be configured
	if err := hc.validateStorage(); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := hc.validateDatabase(); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := hc.validateCache(); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := hc.Spec.ValidateNotary(); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := hc.Spec.ValidateRegistryController(); err != nil {
		allErrs = append(allErrs, err)
	}

	if old == nil { // create harbor resource
		if err := version.Validate(hc.Spec.Version); err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("version"), hc.Spec.Version, err.Error()))
		}
	} else {
		if err := version.UpgradeAllowed(old.Spec.Version, hc.Spec.Version); err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("version"), hc.Spec.Version, err.Error()))
		}
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "HarborCluster"}, hc.Name, allErrs)
}

func (hc *HarborCluster) validateStorage() *field.Error {
	// in cluster storage has high priority
	fp := field.NewPath("spec").Child("inClusterStorage")

	// Storage
	// External is not configured
	if err := hc.Spec.ImageChartStorage.Validate(); err != nil {
		clog.Info("validate spec.imageChartStorage", "cause", err.Error())

		// And in-cluster minIO is not configured
		if hc.Spec.InClusterStorage == nil {
			// Invalid and not acceptable
			return required(fp)
		}
	} else if hc.Spec.InClusterStorage != nil {
		// Both are configured, conflict
		p := field.NewPath("spec").Child("imageChartStorage")

		return forbidden(fp, p)
	}

	// Validate more if incluster storage is configured.
	if hc.Spec.InClusterStorage != nil {
		desiredReplicas := hc.Spec.InClusterStorage.MinIOSpec.Replicas
		volumePerServer := hc.Spec.InClusterStorage.MinIOSpec.VolumesPerServer

		if desiredReplicas*volumePerServer < minimalVolumeCount {
			return invalid(fp, hc.Spec.InClusterStorage.MinIOSpec, fmt.Sprintf("minIOSpec.replicas * minIOSpec.volumesPerServer should be >=%d", minimalVolumeCount))
		}
	}

	return nil
}

func (hc *HarborCluster) validateDatabase() *field.Error {
	// in cluster database has high priority
	fp := field.NewPath("spec").Child("inClusterDatabase")

	// Database
	// External is not configured
	// And also in-cluster psql is not specified
	if hc.Spec.Database == nil && hc.Spec.InClusterDatabase == nil {
		// Invalid and not acceptable
		return required(fp)
	}

	// Both are configured then conflict
	if hc.Spec.Database != nil && hc.Spec.InClusterDatabase != nil {
		p := field.NewPath("spec").Child("database")
		// Conflict and not acceptable
		return forbidden(fp, p)
	}

	return nil
}

func (hc *HarborCluster) validateCache() *field.Error {
	// in cluster cache has high priority
	fp := field.NewPath("spec").Child("inClusterCache")

	// Cache
	// External is not configured
	if hc.Spec.Redis == nil && hc.Spec.InClusterCache == nil {
		// Invalid and not acceptable
		return required(fp)
	}

	// Both are configured and then conflict
	if hc.Spec.Redis != nil && hc.Spec.InClusterCache != nil {
		p := field.NewPath("spec").Child("redis")
		// Conflict and not acceptable
		return forbidden(fp, p)
	}

	return nil
}

func forbidden(mainPath fmt.Stringer, conflictPath *field.Path) *field.Error {
	return field.Forbidden(conflictPath, fmt.Sprintf("conflicts: %s should not be configured as %s has been configured already", conflictPath.String(), mainPath.String()))
}

func required(mainPath *field.Path) *field.Error {
	return field.Required(mainPath, fmt.Sprintf("%s should be configured", mainPath.String()))
}

func invalid(mainPath *field.Path, value interface{}, details string) *field.Error {
	return field.Invalid(mainPath, value, details)
}
