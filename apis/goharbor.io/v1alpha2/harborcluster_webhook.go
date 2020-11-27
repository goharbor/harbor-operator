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

package v1alpha2

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// Log used this webhook
var clog = logf.Log.WithName("harborcluster-resource")

// SetupWebhookWithManager sets up validating webhook of HarborCluster
func (hc *HarborCluster) SetupWebhookWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(hc).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-goharbor-io-v1alpha2-harborcluster,mutating=false,failurePolicy=fail,groups=goharbor.io,resources=harborclusters,versions=v1alpha2,name=vharborcluster.kb.io

var _ webhook.Validator = &HarborCluster{}

func (hc *HarborCluster) ValidateCreate() error {
	clog.Info("validate creation", "name", hc.Name, "namespace", hc.Namespace)

	return hc.validate()
}

func (hc *HarborCluster) ValidateUpdate(old runtime.Object) error {
	clog.Info("validate updating", "name", hc.Name, "namespace", hc.Namespace)

	return hc.validate()
}

func (hc *HarborCluster) ValidateDelete() error {
	clog.Info("validate deletion", "name", hc.Name, "namespace", hc.Namespace)

	return nil
}

func (hc *HarborCluster) validate() error {
	var allErrs field.ErrorList

	// For database(psql), cache(Redis) and storage, either external services or in-cluster services MUST be configured

	// Storage
	// External is not configured
	if err := hc.Spec.ImageChartStorage.Validate(); err != nil {
		// And in-cluster minIO is not configured
		if hc.Spec.InClusterStorage == nil {
			// Invalid and not acceptable
			allErrs = append(
				allErrs,
				field.Invalid(
					field.NewPath("spec").
						Child("imageChartStorage", "inClusterStorage"),
					hc.Spec.ImageChartStorage,
					"both storage and in-cluster storage are not correctly configured",
				),
			)
		}
	}

	// Database
	// External is not configured
	// And also in-cluster psql is not specified
	if hc.Spec.Database == nil && hc.Spec.InClusterDatabase == nil {
		// Invalid and not acceptable
		allErrs = append(
			allErrs,
			field.Invalid(
				field.NewPath("spec").Child("database", "inClusterDatabase"),
				hc.Spec.Database,
				"both database or in-cluster database are not correctly configured",
			),
		)
	}

	// Cache
	// External is not configured
	if hc.Spec.Redis == nil && hc.Spec.InClusterCache == nil {
		// Invalid and not acceptable
		allErrs = append(
			allErrs,
			field.Invalid(
				field.NewPath("spec").Child("redis", "inClusterCache"),
				hc.Spec.Database,
				"both redis or in-cluster redis are not correctly configured",
			),
		)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "HarborCluster"}, hc.Name, allErrs)
}
