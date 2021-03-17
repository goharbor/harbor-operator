package harbor

import (
	"context"
	"time"

	goharborv1alpha2 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha2"
	harbormetav1 "github.com/goharbor/harbor-operator/apis/meta/v1alpha1"
	"github.com/goharbor/harbor-operator/pkg/config"
	commonCtrl "github.com/goharbor/harbor-operator/pkg/controller"
	"github.com/goharbor/harbor-operator/pkg/event-filter/class"
	"github.com/goharbor/harbor-operator/pkg/image"
	certv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	"github.com/ovh/configstore"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

const (
	DefaultRequeueWait = 2 * time.Second
)

// Reconciler reconciles a Harbor object.
type Reconciler struct {
	*commonCtrl.Controller
}

// +kubebuilder:rbac:groups=goharbor.io,resources=harbors,verbs=get;list;watch
// +kubebuilder:rbac:groups=goharbor.io,resources=harbors/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=goharbor.io,resources=chartmuseums;cores;jobservices;notaryservers;notarysigners;portals;registries;registrycontrollers;trivies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=issuers;certificates,verbs=get;list;watch;create;update;patch;delete

func (r *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	err := r.Controller.SetupWithManager(ctx, mgr)
	if err != nil {
		return errors.Wrap(err, "cannot setup common controller")
	}

	className, err := r.GetClassName(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot get class name")
	}

	concurrentReconcile, err := r.ConfigStore.GetItemValueInt(config.ReconciliationKey)
	if err != nil {
		return errors.Wrap(err, "cannot get concurrent reconcile")
	}

	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(&class.Filter{
			ClassName: className,
		}).
		For(r.NewEmpty(ctx)).
		Owns(&goharborv1alpha2.ChartMuseum{}).
		Owns(&goharborv1alpha2.Core{}).
		Owns(&goharborv1alpha2.JobService{}).
		Owns(&goharborv1alpha2.Portal{}).
		Owns(&goharborv1alpha2.Registry{}).
		Owns(&goharborv1alpha2.RegistryController{}).
		Owns(&goharborv1alpha2.NotaryServer{}).
		Owns(&goharborv1alpha2.NotarySigner{}).
		Owns(&corev1.Secret{}).
		Owns(&certv1.Issuer{}).
		Owns(&certv1.Certificate{}).
		Owns(&netv1.Ingress{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: int(concurrentReconcile),
		}).
		Complete(r)
}

func New(ctx context.Context, name string, configStore *configstore.Store) (commonCtrl.Reconciler, error) {
	r := &Reconciler{}

	r.Controller = commonCtrl.NewController(ctx, name, r, configStore)

	return r, nil
}

func (r *Reconciler) getComponentSpec(ctx context.Context, harbor *goharborv1alpha2.Harbor, component harbormetav1.Component) harbormetav1.ComponentSpec {
	var spec harbormetav1.ComponentSpec

	//nolint:golint,exhaustive
	switch component {
	case harbormetav1.ChartMuseumComponent:
		harbor.Spec.ChartMuseum.ComponentSpec.DeepCopyInto(&spec)
	case harbormetav1.CoreComponent:
		harbor.Spec.Core.ComponentSpec.DeepCopyInto(&spec)
	case harbormetav1.JobServiceComponent:
		harbor.Spec.JobService.ComponentSpec.DeepCopyInto(&spec)
	case harbormetav1.NotaryServerComponent:
		harbor.Spec.Notary.Server.DeepCopyInto(&spec)
	case harbormetav1.NotarySignerComponent:
		harbor.Spec.Notary.Signer.DeepCopyInto(&spec)
	case harbormetav1.PortalComponent:
		harbor.Spec.Portal.DeepCopyInto(&spec)
	case harbormetav1.RegistryComponent, harbormetav1.RegistryControllerComponent:
		harbor.Spec.Registry.ComponentSpec.DeepCopyInto(&spec)
	case harbormetav1.TrivyComponent:
		harbor.Spec.Trivy.ComponentSpec.DeepCopyInto(&spec)
	}

	imageSource := harbor.Spec.ImageSource
	if imageSource == nil {
		return spec
	}

	if spec.Image == "" && (imageSource.Repository != "" || imageSource.TagSuffix != "") {
		getImageOptions := []image.Option{
			image.WithRepository(imageSource.Repository),
			image.WithTagSuffix(imageSource.TagSuffix),
			image.WithHarborVersion(harbor.Spec.Version),
		}
		spec.Image, _ = image.GetImage(ctx, component.String(), getImageOptions...)
	}

	if spec.ImagePullPolicy == nil && imageSource.ImagePullPolicy != nil {
		in, out := &imageSource.ImagePullPolicy, &spec.ImagePullPolicy
		*out = new(corev1.PullPolicy)
		**out = **in
	}

	if len(spec.ImagePullSecrets) == 0 && len(imageSource.ImagePullSecrets) != 0 {
		in, out := &imageSource.ImagePullSecrets, &spec.ImagePullSecrets
		*out = make([]corev1.LocalObjectReference, len(*in))
		copy(*out, *in)
	}

	return spec
}
